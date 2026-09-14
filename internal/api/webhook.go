package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/henockt/relay/internal/email"
	"github.com/henockt/relay/internal/models"
)

const (
	maxAttachmentSize = 10 * 1024 * 1024
	replyTokenTTL     = 30 * 24 * time.Hour
	// prepended to forwarded mail, and the cut point when relaying a reply
	forwardedMarker = "--- Forwarded via Relay ---"
	// how much of the multipart body to hold in memory before spilling to disk
	multipartMemory = 8 * 1024 * 1024
)

// handles POST /api/webhooks/email
// called by a Mailgun inbound route when mail arrives
func (s *Server) handleInboundEmail(c *gin.Context) {
	// Mailgun only posts multipart/form-data when the message carries
	// attachments.
	if err := parseInboundForm(c.Request); err != nil {
		log.Printf("webhook: failed to parse inbound email: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	defer func() {
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
	}()

	if !s.verifyMailgunSignature(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	// "recipient" and "sender" are the envelope addresses, the From/To headers
	// can differ and are not what aliases are keyed on.
	to := strings.ToLower(strings.TrimSpace(c.Request.FormValue("recipient")))
	if to == "" {
		log.Printf("webhook: inbound email has no recipient")
		c.Status(http.StatusBadRequest)
		return
	}
	from := strings.TrimSpace(c.Request.FormValue("sender"))
	subject := c.Request.FormValue("subject")
	body := c.Request.FormValue("body-plain")
	if body == "" {
		body = c.Request.FormValue("body-html")
	}
	if body == "" {
		body = "(no body)"
	}

	attachments := collectAttachments(c.Request.MultipartForm, to)

	if replyToken, ok := parseReplyTokenAddress(to, s.cfg.SMTPDomain); ok {
		stripped := c.Request.FormValue("stripped-text")
		s.handleReplyForward(c, replyToken, subject, replyBody(stripped, body), attachments)
		return
	}

	alias, err := s.aliasStore.FindByAddress(to)
	if err != nil {
		log.Printf("webhook: unknown alias %s", to)
		c.Status(http.StatusOK) // intentional so Mailgun does not retry
		return
	}

	if !alias.Enabled {
		alias.EmailsBlocked++
		if err := s.aliasStore.Update(alias); err != nil {
			log.Printf("webhook: failed to update blocked count for alias %s: %v", to, err)
		}
		log.Printf("webhook: alias %s is disabled, blocking", to)
		c.Status(http.StatusOK) // again intentional, a disabled alias is not a delivery failure
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
		"%s\nAlias: %s\nOriginal from: %s\n---\n\n%s",
		forwardedMarker, to, from, body,
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
		OriginalFrom:      from,
		OriginalMessageID: extractMessageID(c.Request),
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

// An anonymous reply must carry only what the user typed. Mail clients quote
// the whole message they are replying to, which here is the forwarded mail
// complete with its relay metadata, so sending the raw body would echo that
// back to the stranger. Mailgun's stripped-text already has quoted history and
// signatures removed; when it is missing, cut the body at our own marker.
func replyBody(stripped, full string) string {
	body := stripped
	if strings.TrimSpace(body) == "" {
		body = full
	}
	body = strings.TrimSpace(truncateAtMarker(body))
	if body == "" {
		return "(no body)"
	}
	return body
}

// the marker usually reappears quoted ("> --- Forwarded via Relay ---"), so
// match anywhere in the line rather than at its start
func truncateAtMarker(body string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.Contains(line, forwardedMarker) {
			return strings.Join(lines[:i], "\n")
		}
	}
	return body
}

func parseInboundForm(r *http.Request) error {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return r.ParseMultipartForm(multipartMemory)
	}
	return r.ParseForm()
}

// Mailgun numbers attachment parts attachment-1, attachment-2, ... Anything
// over maxAttachmentSize is truncated rather than dropping the whole message.
func collectAttachments(form *multipart.Form, to string) []email.Attachment {
	if form == nil {
		return nil
	}

	var attachments []email.Attachment
	for field, headers := range form.File {
		if !strings.HasPrefix(field, "attachment-") {
			continue
		}
		for _, header := range headers {
			file, err := header.Open()
			if err != nil {
				log.Printf("webhook: could not open attachment %s for %s: %v", header.Filename, to, err)
				continue
			}
			data, err := io.ReadAll(io.LimitReader(file, maxAttachmentSize))
			_ = file.Close()
			if err != nil {
				log.Printf("webhook: could not read attachment %s for %s: %v", header.Filename, to, err)
				continue
			}
			attachments = append(attachments, email.Attachment{
				Filename:    header.Filename,
				ContentType: header.Header.Get("Content-Type"),
				Content:     data,
			})
		}
	}

	if len(attachments) > 0 {
		log.Printf("webhook: parsed %d attachment(s) for %s", len(attachments), to)
	}
	return attachments
}

// Mailgun signs each webhook as HMAC-SHA256(timestamp + token) using the
// signing key. Note there is deliberately no timestamp freshness check. Mailgun
// retries a failed webhook for up to 8 hours reusing the original signature, so
// a replay window would reject legitimate retries.
func (s *Server) verifyMailgunSignature(c *gin.Context) bool {
	if s.cfg.MailgunSigningKey == "" {
		log.Printf("webhook: MAILGUN_SIGNING_KEY is not set, refusing inbound email")
		return false
	}

	timestamp := c.Request.FormValue("timestamp")
	token := c.Request.FormValue("token")
	signature := c.Request.FormValue("signature")
	if timestamp == "" || token == "" || signature == "" {
		log.Printf("webhook: inbound email is missing signature fields")
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.MailgunSigningKey))
	mac.Write([]byte(timestamp + token))
	expected := mac.Sum(nil)

	provided, err := hex.DecodeString(signature)
	if err != nil {
		log.Printf("webhook: inbound email has a malformed signature")
		return false
	}
	if !hmac.Equal(expected, provided) {
		log.Printf("webhook: invalid inbound email signature")
		return false
	}
	return true
}

func extractMessageID(r *http.Request) string {
	if messageID := strings.TrimSpace(r.FormValue("Message-Id")); messageID != "" {
		return messageID
	}
	// fall back to the raw MIME headers, which Mailgun sends as a JSON array of
	// [name, value] pairs
	var headers [][]string
	if err := json.Unmarshal([]byte(r.FormValue("message-headers")), &headers); err != nil {
		return ""
	}
	for _, header := range headers {
		if len(header) == 2 && strings.EqualFold(header[0], "Message-Id") {
			return strings.TrimSpace(header[1])
		}
	}
	return ""
}
