package email

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

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

type EmailMessage struct {
	To          string
	From        string
	Subject     string
	Body        string
	ReplyTo     string
	InReplyTo   string
	References  []string
	Attachments []Attachment
}

// interface the webhook handler uses to forward email
type Sender interface {
	Send(message EmailMessage) error
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

func (l *logSender) Send(message EmailMessage) error {
	log.Printf(
		"[EMAIL] to=%s from=%s reply_to=%s subject=%q attachments=%d (log-only, not sent)",
		message.To,
		message.From,
		message.ReplyTo,
		message.Subject,
		len(message.Attachments),
	)
	return nil
}

// sendGridSender sends via the SendGrid SDK
type sendGridSender struct {
	apiKey string
}

func (s *sendGridSender) Send(message EmailMessage) error {
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail("", message.From))
	m.Subject = message.Subject

	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail("", message.To))
	m.AddPersonalizations(p)
	m.AddContent(mail.NewContent("text/plain", message.Body))

	if message.ReplyTo != "" {
		m.SetReplyTo(mail.NewEmail("", message.ReplyTo))
	}
	if message.InReplyTo != "" {
		m.SetHeader("In-Reply-To", message.InReplyTo)
	}
	if len(message.References) > 0 {
		m.SetHeader("References", strings.Join(compact(message.References), " "))
	}

	for _, a := range message.Attachments {
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

func compact(values []string) []string {
	var out []string
	for _, value := range values {
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}
