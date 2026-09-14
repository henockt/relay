package email

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/henockt/relay/internal/config"
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
	if cfg.MailgunAPIKey == "" || cfg.MailgunAPIKey == "dev" {
		return &logSender{}
	}
	return &mailgunSender{
		apiKey:  cfg.MailgunAPIKey,
		domain:  cfg.MailgunDomain,
		baseURL: strings.TrimSuffix(cfg.MailgunAPIBase, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
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

// mailgunSender sends via the Mailgun messages API
type mailgunSender struct {
	apiKey  string
	domain  string
	baseURL string
	client  *http.Client
}

func (s *mailgunSender) Send(message EmailMessage) error {
	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)

	fields := []struct{ key, value string }{
		{"from", message.From},
		{"to", message.To},
		{"subject", message.Subject},
		{"text", message.Body},
		{"h:Reply-To", message.ReplyTo},
		{"h:In-Reply-To", message.InReplyTo},
		{"h:References", strings.Join(compact(message.References), " ")},
	}
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		if err := form.WriteField(f.key, f.value); err != nil {
			return fmt.Errorf("mailgun: write field %s: %w", f.key, err)
		}
	}

	for _, a := range message.Attachments {
		part, err := newAttachmentPart(form, a)
		if err != nil {
			return fmt.Errorf("mailgun: attach %s: %w", a.Filename, err)
		}
		if _, err := part.Write(a.Content); err != nil {
			return fmt.Errorf("mailgun: write attachment %s: %w", a.Filename, err)
		}
	}

	if err := form.Close(); err != nil {
		return fmt.Errorf("mailgun: close form: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", s.baseURL, s.domain)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("mailgun: build request: %w", err)
	}
	req.SetBasicAuth("api", s.apiKey)
	req.Header.Set("Content-Type", form.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailgun send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("mailgun returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// multipart.CreateFormFile always sets application/octet-stream, which loses the
// original type, so build the part by hand to preserve it.
func newAttachmentPart(form *multipart.Writer, a Attachment) (io.Writer, error) {
	filename := a.Filename
	if filename == "" {
		filename = "attachment"
	}
	contentType := a.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(
		`form-data; name="attachment"; filename=%q`, filename,
	))
	header.Set("Content-Type", contentType)
	return form.CreatePart(header)
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
