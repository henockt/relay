package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/henockt/relay/internal/email"
	"github.com/henockt/relay/internal/models"
	"github.com/sendgrid/sendgrid-go/helpers/inbound"
)

const (
	maxAttachmentSize = 10 * 1024 * 1024
	replyTokenTTL     = 30 * 24 * time.Hour
)

// handles POST /api/webhooks/email
// called by SendGrid Inbound Parse when mail arrives
func (s *Server) handleInboundEmail(c *gin.Context) {
	webSecret := c.Query("secret")
	if webSecret != s.cfg.WebhookSecret {
		log.Printf("webhook: invalid secret")
		c.Status(http.StatusUnauthorized)
		return
	}

	parsed, err := inbound.ParseWithAttachments(c.Request)
	if err != nil || len(parsed.Envelope.To) == 0 {
		log.Printf("webhook: failed to parse inbound email: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	// get to, subject, body
	to := strings.ToLower(parsed.Envelope.To[0])
	subject := parsed.ParsedValues["subject"]
	body := parsed.TextBody
	if body == "" {
		body = parsed.Body["text/html"]
	}
	if body == "" {
		body = "(no body)"
	}

	// get attachments
	var attachments []email.Attachment
	for _, a := range parsed.ParsedAttachments {
		data, err := io.ReadAll(io.LimitReader(a.File, maxAttachmentSize))
		if err != nil {
			continue
		}
		attachments = append(attachments, email.Attachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Content:     data,
		})
	}
	if len(attachments) > 0 {
		log.Printf("webhook: parsed %d attachment(s) for %s", len(attachments), to)
	}

	if replyToken, ok := parseReplyTokenAddress(to, s.cfg.SMTPDomain); ok {
		s.handleReplyForward(c, replyToken, subject, body, attachments)
		return
	}

	alias, err := s.aliasStore.FindByAddress(to)
	if err != nil {
		log.Printf("webhook: unknown alias %s", to)
		c.Status(http.StatusOK) // intentional so SendGrid doesn't retry
		return
	}

	if !alias.Enabled {
		alias.EmailsBlocked++
		if err := s.aliasStore.Update(alias); err != nil {
			log.Printf("webhook: failed to update blocked count for alias %s: %v", to, err)
		}
		log.Printf("webhook: alias %s is disabled, blocking", to)
		c.Status(http.StatusOK) // again intentional
		return
	}

	user, err := s.userStore.FindByID(alias.UserID)
	if err != nil {
		log.Printf("webhook: user not found for alias %s: %v", to, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Prepend relay metadata to the body so the user knows which alias received it.
	forwardedBody := fmt.Sprintf(
		"--- Forwarded via Relay ---\nAlias: %s\nOriginal from: %s\n---\n\n%s",
		to, parsed.Envelope.From, body,
	)

	replyToken, err := generateReplyToken()
	if err != nil {
		log.Printf("webhook: failed to generate reply token for alias %s: %v", to, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	replyThread := &models.ReplyThread{
		ReplyToken:        replyToken,
		AliasID:           alias.ID,
		OriginalFrom:      parsed.Envelope.From,
		OriginalMessageID: extractMessageID(parsed),
		ExpiresAt:         time.Now().Add(replyTokenTTL),
	}
	if err := s.replyThreadStore.Create(replyThread); err != nil {
		log.Printf("webhook: failed to persist reply thread for alias %s: %v", to, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	replyToAddress := fmt.Sprintf("r+%s@%s", replyToken, s.cfg.SMTPDomain)
	fromAddr := fmt.Sprintf("relay+%s@%s", strings.Split(to, "@")[0], s.cfg.SMTPDomain)
	if err := s.sender.Send(email.EmailMessage{
		To:          user.Email,
		From:        fromAddr,
		Subject:     subject,
		Body:        forwardedBody,
		ReplyTo:     replyToAddress,
		Attachments: attachments,
	}); err != nil {
		alias.EmailsBlocked++
		if err := s.aliasStore.Update(alias); err != nil {
			log.Printf("webhook: failed to update blocked count for alias %s: %v", to, err)
		}
		log.Printf("webhook: forward failed for alias %s: %v", to, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	alias.EmailsForwarded++
	if err := s.aliasStore.Update(alias); err != nil {
		log.Printf("webhook: failed to update forwarded count for alias %s: %v", to, err)
	}
	log.Printf("webhook: forwarded mail for alias %s -> %s", to, user.Email)
	c.Status(http.StatusOK)
}

func (s *Server) handleReplyForward(c *gin.Context, replyToken, subject, body string, attachments []email.Attachment) {
	replyThread, err := s.replyThreadStore.FindByToken(replyToken)
	if err != nil {
		log.Printf("webhook: unknown reply token %s", replyToken)
		c.Status(http.StatusOK)
		return
	}

	if time.Now().After(replyThread.ExpiresAt) {
		log.Printf("webhook: expired reply token %s", replyToken)
		c.Status(http.StatusOK)
		return
	}

	alias, err := s.aliasStore.FindByID(replyThread.AliasID)
	if err != nil {
		log.Printf("webhook: alias not found for reply token %s: %v", replyToken, err)
		c.Status(http.StatusOK)
		return
	}

	if !alias.Enabled {
		alias.EmailsBlocked++
		if err := s.aliasStore.Update(alias); err != nil {
			log.Printf("webhook: failed to update blocked count for alias %s: %v", alias.Address, err)
		}
		log.Printf("webhook: blocked reply for disabled alias %s", alias.Address)
		c.Status(http.StatusOK)
		return
	}

	message := email.EmailMessage{
		To:          replyThread.OriginalFrom,
		From:        alias.Address,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
	}
	if replyThread.OriginalMessageID != "" {
		message.InReplyTo = replyThread.OriginalMessageID
		message.References = []string{replyThread.OriginalMessageID}
	}

	if err := s.sender.Send(message); err != nil {
		alias.EmailsBlocked++
		if updateErr := s.aliasStore.Update(alias); updateErr != nil {
			log.Printf("webhook: failed to update blocked count for alias %s: %v", alias.Address, updateErr)
		}
		log.Printf("webhook: failed to relay anonymous reply for alias %s: %v", alias.Address, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	alias.EmailsForwarded++
	if err := s.aliasStore.Update(alias); err != nil {
		log.Printf("webhook: failed to update forwarded count for alias %s: %v", alias.Address, err)
	}

	log.Printf("webhook: relayed anonymous reply via alias %s -> %s", alias.Address, replyThread.OriginalFrom)
	c.Status(http.StatusOK)
}

func generateReplyToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate reply token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func parseReplyTokenAddress(address, domain string) (string, bool) {
	parts := strings.SplitN(strings.ToLower(address), "@", 2)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[1], domain) {
		return "", false
	}
	if !strings.HasPrefix(parts[0], "r+") {
		return "", false
	}
	token := strings.TrimPrefix(parts[0], "r+")
	if token == "" {
		return "", false
	}
	return token, true
}

func extractMessageID(parsed *inbound.ParsedEmail) string {
	if messageID := findHeader(parsed.Headers, "Message-ID"); messageID != "" {
		return messageID
	}
	if messageID := findHeader(parsed.ParsedValues, "Message-ID"); messageID != "" {
		return messageID
	}
	if messageID := findHeader(parsed.ParsedValues, "Message-Id"); messageID != "" {
		return messageID
	}
	return ""
}

func findHeader(values map[string]string, name string) string {
	for key, value := range values {
		if strings.EqualFold(key, name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
