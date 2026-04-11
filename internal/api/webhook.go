package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/henockt/relay/internal/email"
	"github.com/sendgrid/sendgrid-go/helpers/inbound"
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

	// get attachments
	var attachments []email.Attachment
	for _, a := range parsed.ParsedAttachments {
		data, err := io.ReadAll(a.File)
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

	fromAddr := fmt.Sprintf("relay+%s@%s", strings.Split(to, "@")[0], s.cfg.SMTPDomain)
	if err := s.sender.Send(user.Email, fromAddr, subject, forwardedBody, attachments); err != nil {
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
