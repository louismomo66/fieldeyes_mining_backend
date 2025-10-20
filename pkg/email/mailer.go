package email

import (
	"fmt"
	"log"
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
