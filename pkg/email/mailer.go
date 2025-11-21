package email

import (
	"fmt"
	"log"
	"net/smtp"
)

// Mailer interface for sending emails
type Mailer interface {
	SendOTP(email, otp string) error
}

// MockMailer is a mock implementation for development
type MockMailer struct{}

// SendOTP sends an OTP email (mock implementation)
func (m *MockMailer) SendOTP(email, otp string) error {
	log.Printf("Mock email sent to %s with OTP: %s", email, otp)
	fmt.Printf("📧 Mock Email to %s: Your OTP is %s\n", email, otp)
	return nil
}

// SMTPMailer sends emails using SMTP
type SMTPMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSMTPMailer creates a new SMTP mailer instance
func NewSMTPMailer(host string, port int, username, password, from string) *SMTPMailer {
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// SendOTP sends an OTP via SMTP
func (m *SMTPMailer) SendOTP(email, otp string) error {
	msg := fmt.Sprintf("Subject: Mine Manager OTP\r\n"+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"+
		"Your password reset code is %s.", m.from, email, otp)

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	return smtp.SendMail(addr, auth, m.from, []string{email}, []byte(msg))
}
