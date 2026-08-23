package validator

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
	phoneRegex   = regexp.MustCompile(`^\+?[0-9\s\-()]{7,25}$`)
)

// SanitizeText removes potential HTML/script markup and strips surrounding whitespace.
// Why: Prevents Cross-Site Scripting (XSS) and malformed control characters from being stored in the database.
func SanitizeText(input string) string {
	cleaned := htmlTagRegex.ReplaceAllString(input, "")
	return strings.TrimSpace(cleaned)
}

// ValidateUUID checks if a string conforms to standard RFC 4122 UUID formatting.
// Why: Enforces valid foreign key and identity references across services.
func ValidateUUID(id string) bool {
	_, err := uuid.Parse(strings.TrimSpace(id))
	return err == nil
}

// ValidatePhoneNumber verifies if a string is a valid contact telephone format.
// Why: Validates phone numbers before persisting to prevent corrupted courier / notification records.
func ValidatePhoneNumber(phone string) bool {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return false
	}
	return phoneRegex.MatchString(trimmed)
}

// ValidateGender checks if a provided gender string matches accepted values.
// Why: Normalizes categorical profile demographics.
func ValidateGender(gender string) bool {
	switch strings.ToUpper(strings.TrimSpace(gender)) {
	case "MALE", "FEMALE", "OTHER", "PREFER_NOT_TO_SAY":
		return true
	default:
		return false
	}
}
