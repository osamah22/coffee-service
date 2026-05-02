package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"
)

type Email struct {
	To      string
	Subject string
	Text    string
}

type EmailSender interface {
	Send(ctx context.Context, email Email) error
}

type LogEmailSender struct{}

func NewLogEmailSender() *LogEmailSender {
	return &LogEmailSender{}
}

func (s *LogEmailSender) Send(_ context.Context, email Email) error {
	log.Printf("email dry-run to=%s subject=%q\n%s", email.To, email.Subject, email.Text)
	return nil
}

type SMTPEmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPEmailSender(host string, port int, username, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPEmailSender) Send(ctx context.Context, email Email) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	to := strings.TrimSpace(email.To)
	if to == "" {
		return fmt.Errorf("email recipient is required")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.Auth(nil)
	if s.username != "" || s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	msg := buildMessage(s.from, to, email.Subject, email.Text)
	from, err := envelopeAddress(s.from)
	if err != nil {
		return err
	}

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func buildMessage(from, to, subject, body string) []byte {
	var msg bytes.Buffer
	headers := map[string]string{
		"From":                      from,
		"To":                        to,
		"Subject":                   mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version":              "1.0",
		"Content-Type":              `text/plain; charset="utf-8"`,
		"Content-Transfer-Encoding": "8bit",
	}

	for key, value := range headers {
		msg.WriteString(key)
		msg.WriteString(": ")
		msg.WriteString(value)
		msg.WriteString("\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)
	msg.WriteString("\r\n")
	return msg.Bytes()
}

func envelopeAddress(value string) (string, error) {
	address, err := mail.ParseAddress(value)
	if err != nil {
		return strings.TrimSpace(value), nil
	}
	return address.Address, nil
}
