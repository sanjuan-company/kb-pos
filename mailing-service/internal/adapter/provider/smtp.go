package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"mailing-service/internal/domain"
)

type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromEmail  string
	FromName   string
}

type SMTPProvider struct {
	config SMTPConfig
}

func NewSMTP(cfg SMTPConfig) *SMTPProvider {
	return &SMTPProvider{config: cfg}
}

func (p *SMTPProvider) Name() string {
	return "smtp"
}

func (p *SMTPProvider) Send(ctx context.Context, email *domain.Email) error {
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	msg, err := p.buildMessage(email)
	if err != nil {
		return fmt.Errorf("build message: %w", err)
	}

	client, err := p.dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	from := p.config.FromEmail
	if email.From.Email != "" {
		from = email.From.Email
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}

	for _, addr := range email.To {
		if err := client.Rcpt(addr.Email); err != nil {
			return fmt.Errorf("rcpt %s: %w", addr.Email, err)
		}
	}
	for _, addr := range email.Cc {
		if err := client.Rcpt(addr.Email); err != nil {
			return fmt.Errorf("rcpt %s: %w", addr.Email, err)
		}
	}
	for _, addr := range email.Bcc {
		if err := client.Rcpt(addr.Email); err != nil {
			return fmt.Errorf("rcpt %s: %w", addr.Email, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return client.Quit()
}

func (p *SMTPProvider) HealthCheck(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (p *SMTPProvider) dial(ctx context.Context, addr string) (*smtp.Client, error) {
	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("tcp dial: %w", err)
	}

	wc := &deadlineConn{Conn: conn, ctx: ctx}

	if p.config.Port == 465 {
		tlsConn := tls.Client(wc, &tls.Config{ServerName: p.config.Host})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			wc.Close()
			return nil, fmt.Errorf("tls handshake: %w", err)
		}
		return smtp.NewClient(tlsConn, p.config.Host)
	}

	client, err := smtp.NewClient(wc, p.config.Host)
	if err != nil {
		wc.Close()
		return nil, fmt.Errorf("smtp client: %w", err)
	}
	if err := client.StartTLS(&tls.Config{ServerName: p.config.Host}); err != nil {
		client.Close()
		return nil, fmt.Errorf("starttls: %w", err)
	}
	return client, nil
}

type deadlineConn struct {
	net.Conn
	ctx context.Context
}

func (c *deadlineConn) Read(b []byte) (int, error) {
	if err := c.setDeadline(); err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *deadlineConn) Write(b []byte) (int, error) {
	if err := c.setDeadline(); err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

func (c *deadlineConn) setDeadline() error {
	if deadline, ok := c.ctx.Deadline(); ok {
		c.Conn.SetDeadline(deadline)
	}
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
		return nil
	}
}

func (p *SMTPProvider) buildMessage(email *domain.Email) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("From: %s\r\n", p.formatAddress(p.resolveFrom(email))))
	b.WriteString(fmt.Sprintf("To: %s\r\n", p.formatAddresses(email.To)))
	if len(email.Cc) > 0 {
		b.WriteString(fmt.Sprintf("Cc: %s\r\n", p.formatAddresses(email.Cc)))
	}
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	b.WriteString("MIME-Version: 1.0\r\n")

	if email.HTML != "" {
		b.WriteString("Content-Type: multipart/alternative; boundary=\"boundary-123\"\r\n")
		b.WriteString("\r\n")
		b.WriteString("--boundary-123\r\n")
		b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(email.PlainText)
		b.WriteString("\r\n")
		b.WriteString("--boundary-123\r\n")
		b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(email.HTML)
		b.WriteString("\r\n")
		b.WriteString("--boundary-123--\r\n")
	} else {
		b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(email.PlainText)
	}

	return b.String(), nil
}

func (p *SMTPProvider) resolveFrom(email *domain.Email) domain.Address {
	if email.From.Email != "" {
		return email.From
	}
	return domain.Address{
		Email: p.config.FromEmail,
		Name:  p.config.FromName,
	}
}

func (p *SMTPProvider) formatAddress(addr domain.Address) string {
	if addr.Name == "" {
		return addr.Email
	}
	return fmt.Sprintf("%s <%s>", addr.Name, addr.Email)
}

func (p *SMTPProvider) formatAddresses(addrs []domain.Address) string {
	parts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		parts = append(parts, p.formatAddress(a))
	}
	return strings.Join(parts, ", ")
}
