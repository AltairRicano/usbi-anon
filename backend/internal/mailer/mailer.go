// Package mailer sends transactional email through a generic SMTP server.
//
// It intentionally depends only on the standard library (net/smtp, crypto/tls)
// so no third-party mail provider is required — connection details come from the
// environment (SMTP_HOST/PORT/USER/PASSWORD/FROM). A LogMailer fallback is used
// in local development when no SMTP server is configured.
package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// ErrNotConfigured is returned by a mailer that has no usable transport.
var ErrNotConfigured = errors.New("mailer: not configured")

// Mailer sends a single plain-text email. Implementations must be safe for
// concurrent use by multiple goroutines.
type Mailer interface {
	Send(ctx context.Context, to, subject, textBody string) error
}

// Config holds SMTP connection settings, typically loaded from the environment.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	// ImplicitTLS dials a TLS connection immediately (SMTPS, usually port 465).
	// When false, the client upgrades opportunistically via STARTTLS if the
	// server advertises it.
	ImplicitTLS bool
	// Timeout bounds the whole SMTP exchange. Defaults to 20s.
	Timeout time.Duration
}

// SMTPMailer sends mail through a generic SMTP server using the standard library.
type SMTPMailer struct {
	cfg Config
}

// NewSMTPMailer builds an SMTPMailer. It does not connect until Send is called.
func NewSMTPMailer(cfg Config) *SMTPMailer {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 20 * time.Second
	}
	return &SMTPMailer{cfg: cfg}
}

// Send delivers a plain-text UTF-8 message to a single recipient. It negotiates
// STARTTLS (or implicit TLS) and authenticates with PLAIN when credentials are
// present. Auth is never attempted over an un-encrypted connection.
func (m *SMTPMailer) Send(ctx context.Context, to, subject, textBody string) error {
	if m.cfg.Host == "" {
		return ErrNotConfigured
	}
	from := m.cfg.From
	if from == "" {
		from = m.cfg.Username
	}
	if from == "" {
		return fmt.Errorf("mailer: no From address configured")
	}

	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	msg := buildMessage(from, to, subject, textBody)

	dialer := net.Dialer{Timeout: m.cfg.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("mailer: dial %s: %w", addr, err)
	}
	// A single deadline bounds the entire exchange (net/smtp has no context API).
	deadline := time.Now().Add(m.cfg.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)

	tlsCfg := &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}
	if m.cfg.ImplicitTLS {
		conn = tls.Client(conn, tlsCfg)
	}

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("mailer: smtp handshake: %w", err)
	}
	defer func() { _ = client.Close() }()

	if !m.cfg.ImplicitTLS {
		// STARTTLS is mandatory, not opportunistic: a MITM that strips the
		// STARTTLS capability from the server's EHLO must NOT be able to
		// downgrade this exchange to plaintext (CN-002 / STARTTLS stripping).
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("mailer: server does not advertise STARTTLS; refusing to send over an unencrypted connection")
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("mailer: starttls: %w", err)
		}
	}

	// Defence-in-depth: never transmit credentials or the message body unless
	// the connection is actually encrypted — an implicit-TLS dial or a completed
	// STARTTLS upgrade. Guards against ever sending AUTH PLAIN in cleartext.
	if _, isTLS := client.TLSConnectionState(); !isTLS {
		return fmt.Errorf("mailer: connection is not encrypted; refusing to authenticate or send")
	}

	if m.cfg.Username != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("mailer: auth: %w", err)
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("mailer: RCPT TO: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := wc.Write([]byte(msg)); err != nil {
		_ = wc.Close()
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}
	return client.Quit()
}

// buildMessage renders a minimal RFC 5322 plain-text message with CRLF line
// endings and a MIME-encoded subject (so accented Spanish text survives).
func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	// Normalise to CRLF as required by SMTP.
	b.WriteString(strings.ReplaceAll(body, "\n", "\r\n"))
	return b.String()
}

// LogMailer writes messages to a logger instead of sending them. It exists only
// for local development where no SMTP server is configured; it must NEVER be
// used in production, because the verification link would only reach the logs.
type LogMailer struct {
	logger *log.Logger
}

// NewLogMailer builds a LogMailer around the given logger (defaults to the
// standard logger when nil).
func NewLogMailer(logger *log.Logger) *LogMailer {
	if logger == nil {
		logger = log.Default()
	}
	return &LogMailer{logger: logger}
}

// Send records the message to the log and always succeeds.
func (m *LogMailer) Send(_ context.Context, to, subject, textBody string) error {
	m.logger.Printf("[MAILER:DEV] would send email to=%q subject=%q\n---\n%s\n---", to, subject, textBody)
	return nil
}
