package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"app/internal/config"
)

// SMTPMailer sends emails via SMTP.
type SMTPMailer struct {
	cfg config.SMTPConfig
	log *slog.Logger
}

// NewSMTPMailer creates an SMTP-backed email sender.
func NewSMTPMailer(cfg config.SMTPConfig, log *slog.Logger) *SMTPMailer {
	return &SMTPMailer{cfg: cfg, log: log}
}

// SendPasswordReset sends a password reset link to the user.
func (m *SMTPMailer) SendPasswordReset(ctx context.Context, to, resetLink string) error {
	_ = ctx

	subject := "Reset your password"
	body := fmt.Sprintf(
		"Hello,\n\nWe received a request to reset your password. Click the link below to choose a new password:\n\n%s\n\nIf you did not request this, you can safely ignore this email.\n\nThis link expires in a short time.\n",
		resetLink,
	)

	if !m.cfg.Enabled {
		m.log.Info("smtp disabled — password reset link",
			slog.String("to", to),
			slog.String("reset_link", resetLink),
		)
		return nil
	}

	from := m.cfg.FromEmail
	if m.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.FromEmail)
	}

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if m.cfg.UseTLS {
		return m.sendTLS(addr, auth, m.cfg.FromEmail, []string{to}, []byte(msg))
	}

	return smtp.SendMail(addr, auth, m.cfg.FromEmail, []string{to}, []byte(msg))
}

func (m *SMTPMailer) sendTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host := m.cfg.Host

	tlsConfig := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if auth != nil && m.cfg.Username != "" {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}

	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}
