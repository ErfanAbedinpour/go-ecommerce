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
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>You recently requested to reset your password. Click the link below to proceed:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>If you did not request this, please ignore this email.</p>
	`, resetLink)

	return m.send(to, subject, body)
}

func (m *SMTPMailer) SendOrderConfirmation(ctx context.Context, to string, orderNumber string, total float64) error {
	subject := fmt.Sprintf("Order Confirmation - %s", orderNumber)
	body := fmt.Sprintf(`
		<h2>Thank you for your order!</h2>
		<p>Your order <strong>%s</strong> has been received and is being processed.</p>
		<p>Order Total: $%.2f</p>
		<p>We will notify you when it ships.</p>
	`, orderNumber, total)

	return m.send(to, subject, body)
}

func (m *SMTPMailer) send(to, subject, body string) error {
	if !m.cfg.Enabled {
		m.log.Info("smtp disabled — sending email",
			slog.String("to", to),
			slog.String("subject", subject),
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
