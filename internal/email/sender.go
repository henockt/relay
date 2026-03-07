package email

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/henockt/relay/internal/config"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Attachment holds a single email attachment to be forwarded
type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
}

// interface the webhook handler uses to forward email
type Sender interface {
	Send(to, from, subject, body string, attachments []Attachment) error
}

// returns the right implementation based on config
func NewSender(cfg *config.Config) Sender {
	if cfg.SendGridAPIKey == "" || cfg.SendGridAPIKey == "dev" {
		return &logSender{}
	}
	return &sendGridSender{apiKey: cfg.SendGridAPIKey}
}

// logSender prints to stdout
type logSender struct{}

func (l *logSender) Send(to, from, subject, body string, attachments []Attachment) error {
	log.Printf("[EMAIL] to=%s from=%s subject=%q attachments=%d (log-only, not sent)", to, from, subject, len(attachments))
	return nil
}

// sendGridSender sends via the SendGrid SDK
type sendGridSender struct {
	apiKey string
}

func (s *sendGridSender) Send(to, from, subject, body string, attachments []Attachment) error {
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail("", from))
	m.Subject = subject

	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail("", to))
	m.AddPersonalizations(p)
	m.AddContent(mail.NewContent("text/plain", body))

	for _, a := range attachments {
		att := mail.NewAttachment()
		att.SetContent(base64.StdEncoding.EncodeToString(a.Content))
		att.SetType(a.ContentType)
		att.SetFilename(a.Filename)
		m.AddAttachment(att)
	}

	client := sendgrid.NewSendClient(s.apiKey)
	resp, err := client.Send(m)
	if err != nil {
		return fmt.Errorf("sendgrid send: %w", err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid returned %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}
