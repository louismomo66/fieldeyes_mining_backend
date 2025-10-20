package utils

import (
	"regexp"
	"strings"
)

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword validates password strength
func ValidatePassword(password string) bool {
	return len(password) >= 6
}

// ValidateRequired validates required fields
func ValidateRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) bool {
	if phone == "" {
		return true // Phone is optional
	}
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

// ValidatePositiveNumber validates that a number is positive
func ValidatePositiveNumber(value float64) bool {
	return value > 0
}

// ValidateNonNegativeNumber validates that a number is non-negative
func ValidateNonNegativeNumber(value float64) bool {
	return value >= 0
}
