package services

import (
	"net/mail"
	"strings"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	parsed, err := mail.ParseAddress(email)
	return err == nil && parsed.Address == email
}

func validName(value string) bool {
	value = strings.TrimSpace(value)
	return len(value) >= 2 && len(value) <= 255
}

func validType(value string) bool {
	value = strings.TrimSpace(value)
	return len(value) >= 2 && len(value) <= 50
}
