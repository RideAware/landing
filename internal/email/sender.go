package email

import (
  "fmt"

  "github.com/wneessen/go-mail"
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
  client, err := mail.NewClient(
    s.cfg.SMTPHost,
    mail.WithPort(587),
    mail.WithSMTPAuth(mail.SMTPAuthPlain),
    mail.WithUsername(s.cfg.SMTPUser),
    mail.WithPassword(s.cfg.SMTPPass),
  )
  if err != nil {
    return fmt.Errorf("failed to create mail client: %w", err)
  }

  msg := mail.NewMsg()
  if err := msg.From(s.cfg.SMTPUser); err != nil {
    return fmt.Errorf("failed to set from: %w", err)
  }
  if err := msg.To(email); err != nil {
    return fmt.Errorf("failed to set to: %w", err)
  }

  msg.Subject("Thanks for subscribing!")

  htmlBody := fmt.Sprintf(`
    <html>
    <body>
      <h1>Welcome!</h1>
      <p>Thank you for subscribing to our newsletter.</p>
      <p><a href="%s">Unsubscribe</a></p>
    </body>
    </html>
  `, unsubscribeLink)

  msg.SetBodyString(mail.TypeTextHTML, htmlBody)

  if err := client.DialAndSend(msg); err != nil {
    return fmt.Errorf("failed to send email: %w", err)
  }

  return nil
}