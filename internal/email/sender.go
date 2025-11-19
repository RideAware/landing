package email

import (
  "crypto/tls"
  "fmt"
  "html"
  "net/smtp"
  "strconv"
  "strings"

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

  return s.sendEmail(email, subject, htmlBody)
}

func (s *Sender) SendContactConfirmation(email, name string) error {
  subject := "We received your message - RideAware"
  htmlBody := fmt.Sprintf(`
    <html>
    <body>
      <h2>Thank you for reaching out, %s!</h2>
      <p>We've received your message and will get back to you as soon as possible.</p>
      <p>In the meantime, feel free to check out more about RideAware on our website.</p>
      <p>Best regards,<br>The RideAware Team</p>
    </body>
    </html>
  `, html.EscapeString(name))

  return s.sendEmail(email, subject, htmlBody)
}

func (s *Sender) SendContactNotification(
  adminEmail, name, email, subject, message string,
) error {
  emailSubject := fmt.Sprintf("New contact message from %s", name)
  htmlBody := fmt.Sprintf(`
    <html>
    <body>
      <h3>New Contact Message</h3>
      <p><strong>From:</strong> %s (%s)</p>
      <p><strong>Subject:</strong> %s</p>
      <h4>Message:</h4>
      <p>%s</p>
    </body>
    </html>
  `,
    html.EscapeString(name),
    html.EscapeString(email),
    html.EscapeString(subject),
    strings.ReplaceAll(html.EscapeString(message), "\n", "<br>"),
  )

  return s.sendEmail(adminEmail, emailSubject, htmlBody)
}

func (s *Sender) sendEmail(toEmail, subject, htmlBody string) error {
  port, err := strconv.Atoi(s.cfg.SMTPPort)
  if err != nil {
    return fmt.Errorf("invalid SMTP port '%s': %w", s.cfg.SMTPPort, err)
  }

  message := fmt.Sprintf(
    "From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
    s.cfg.SMTPUser,
    toEmail,
    subject,
    htmlBody,
  )

  addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, port)

  // Port 465 uses direct SSL/TLS
  return s.sendEmailSSL(addr, toEmail, message)
}

func (s *Sender) sendEmailSSL(addr, toEmail, message string) error {
  // Create TLS config
  tlsConfig := &tls.Config{
    ServerName: s.cfg.SMTPHost,
  }

  // Try to dial with TLS
  conn, err := tls.Dial("tcp", addr, tlsConfig)
  if err != nil {
    return fmt.Errorf("failed to dial TLS to %s: %w", addr, err)
  }
  defer conn.Close()

  // Create SMTP client
  client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
  if err != nil {
    return fmt.Errorf("failed to create SMTP client: %w", err)
  }
  defer client.Close()

  // Authenticate
  auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
  if err := client.Auth(auth); err != nil {
    return fmt.Errorf("failed to authenticate with %s: %w", s.cfg.SMTPUser, err)
  }

  // Set sender
  if err := client.Mail(s.cfg.SMTPUser); err != nil {
    return fmt.Errorf("failed to set mail from %s: %w", s.cfg.SMTPUser, err)
  }

  // Set recipient
  if err := client.Rcpt(toEmail); err != nil {
    return fmt.Errorf("failed to set mail to %s: %w", toEmail, err)
  }

  // Get data writer
  wc, err := client.Data()
  if err != nil {
    return fmt.Errorf("failed to get data writer: %w", err)
  }
  defer wc.Close()

  // Write message
  if _, err := wc.Write([]byte(message)); err != nil {
    return fmt.Errorf("failed to write message: %w", err)
  }

  // Quit - ignore quit errors since email was already queued
  _ = client.Quit()

  return nil
}

func (s *Sender) TestConnection() error {
  port, err := strconv.Atoi(s.cfg.SMTPPort)
  if err != nil {
    return fmt.Errorf("invalid SMTP port '%s': %w", s.cfg.SMTPPort, err)
  }

  addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, port)

  // Test TLS connection
  tlsConfig := &tls.Config{
    ServerName: s.cfg.SMTPHost,
  }

  conn, err := tls.Dial("tcp", addr, tlsConfig)
  if err != nil {
    return fmt.Errorf("failed to dial TLS to %s: %w", addr, err)
  }
  defer conn.Close()

  // Test SMTP client creation
  client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
  if err != nil {
    return fmt.Errorf("failed to create SMTP client: %w", err)
  }
  defer client.Close()

  return nil
}