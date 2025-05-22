package notifications

import (
	"gopkg.in/gomail.v2"
)

// EmailSender defines methods for sending emails.
type EmailSender interface {
	SendEmail(to, subject, body string) error
}

// SMTPSender implements EmailSender using SMTP.
type SMTPSender struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
}

func NewSMTPSender(host string, port int, username, password, from string) *SMTPSender {
	return &SMTPSender{
		SMTPHost: host,
		SMTPPort: port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *SMTPSender) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	dialer := gomail.NewDialer(s.SMTPHost, s.SMTPPort, s.Username, s.Password)
	return dialer.DialAndSend(m)
}
