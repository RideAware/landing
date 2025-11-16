package email

import (
  "crypto/tls"
  "fmt"
  "net"
  "net/smtp"
  "strconv"
  "time"

  "landing/internal/config"
)

type Sender struct {
  cfg *config.Config
}

func New(cfg *config.Config) *Sender {
  return &Sender{cfg: cfg}
}

func (s *Sender) SendConfirmationEmail(
  email string,
  unsubscribeLink string,
) error {
  // Parse SMTP port from env
  port, err := strconv.Atoi(s.cfg.SMTPPort)
  if err != nil {
    return fmt.Errorf("invalid SMTP port '%s': %w", s.cfg.SMTPPort, err)
  }

  // Build email message
  subject := "Thanks for subscribing!"
  htmlBody := fmt.Sprintf(`
    <html>
    <body>
      <h1>Welcome to RideAware!</h1>
      <p>Thank you for subscribing to our newsletter.</p>
      <p><a href="%s">Unsubscribe</a></p>
    </body>
    </html>
  `, unsubscribeLink)

  message := fmt.Sprintf(
    "From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
    s.cfg.SMTPUser,
    email,
    subject,
    htmlBody,
  )

  // Create TLS config
  tlsConfig := &tls.Config{
    ServerName: s.cfg.SMTPHost,
  }

  // Send email using smtp.SendMail
  addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, port)
  auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)

  // Use a custom dialer with timeout
  conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
  if err != nil {
    return fmt.Errorf("failed to connect to SMTP server: %w", err)
  }
  defer conn.Close()

  client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
  if err != nil {
    return fmt.Errorf("failed to create SMTP client: %w", err)
  }
  defer client.Close()

  // Start TLS
  if err := client.StartTLS(tlsConfig); err != nil {
    return fmt.Errorf("failed to start TLS: %w", err)
  }

  // Authenticate
  if err := client.Auth(auth); err != nil {
    return fmt.Errorf("failed to authenticate: %w", err)
  }

  // Set recipient and send
  if err := client.Mail(s.cfg.SMTPUser); err != nil {
    return fmt.Errorf("failed to set mail from: %w", err)
  }

  if err := client.Rcpt(email); err != nil {
    return fmt.Errorf("failed to set mail to: %w", err)
  }

  wc, err := client.Data()
  if err != nil {
    return fmt.Errorf("failed to get data writer: %w", err)
  }
  defer wc.Close()

  if _, err := wc.Write([]byte(message)); err != nil {
    return fmt.Errorf("failed to write message: %w", err)
  }

  if err := client.Quit(); err != nil {
    return fmt.Errorf("failed to quit SMTP: %w", err)
  }

  return nil
}

// TestConnection tests SMTP connection without sending email
func (s *Sender) TestConnection() error {
  port, err := strconv.Atoi(s.cfg.SMTPPort)
  if err != nil {
    return fmt.Errorf("invalid SMTP port '%s': %w", s.cfg.SMTPPort, err)
  }

  // Test TCP connection
  addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, port)
  conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
  if err != nil {
    return fmt.Errorf("TCP connection failed to %s: %w", addr, err)
  }
  defer conn.Close()

  // Test SMTP connection
  client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
  if err != nil {
    return fmt.Errorf("failed to create SMTP client: %w", err)
  }
  defer client.Close()

  return nil
}