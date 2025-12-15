// Package validator provides utility functions for data validation.
package validator

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// IsEmail validates if a string is a valid email address.
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsURL validates if a string is a valid URL.
func IsURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsIP validates if a string is a valid IP address (IPv4 or IPv6).
func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

// IsIPv4 validates if a string is a valid IPv4 address.
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

// IsIPv6 validates if a string is a valid IPv6 address.
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}

// IsNumeric checks if a string contains only numeric characters.
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsInteger checks if a string is a valid integer.
func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// IsFloat checks if a string is a valid float.
func IsFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsAlpha checks if a string contains only alphabetic characters.
func IsAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsAlphanumeric checks if a string contains only alphabetic and numeric characters.
func IsAlphanumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsPhone validates a phone number (basic validation).
func IsPhone(phone string) bool {
	// Remove common formatting characters
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if remaining characters are digits and length is reasonable
	if len(cleaned) < 7 || len(cleaned) > 15 {
		return false
	}

	for _, r := range cleaned {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsCreditCard validates a credit card number using the Luhn algorithm.
func IsCreditCard(cardNumber string) bool {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(cardNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	// Check if all characters are digits
	for _, r := range cleaned {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	// Apply Luhn algorithm
	return luhnCheck(cleaned)
}

// luhnCheck implements the Luhn algorithm for credit card validation.
func luhnCheck(cardNumber string) bool {
	if len(cardNumber) < 2 {
		return false
	}

	sum := 0
	alternate := false

	// Start from the rightmost digit
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(cardNumber[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// IsPasswordStrong checks if a password meets basic strength requirements.
// Requirements: at least 8 characters, contains uppercase, lowercase, digit, and special character.
func IsPasswordStrong(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsHexColor validates if a string is a valid hex color code.
func IsHexColor(color string) bool {
	hexColorRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	return hexColorRegex.MatchString(color)
}

// IsUUID validates if a string is a valid UUID.
func IsUUID(s string) bool {
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
	return uuidRegex.MatchString(s)
}

// IsJSON validates if a string is valid JSON.
func IsJSON(s string) bool {
	return regexp.MustCompile(`^\s*[\[\{]`).MatchString(s) &&
		len(strings.TrimSpace(s)) > 0
}

// InRange checks if a number (as string) is within the specified range (inclusive).
func InRange(s string, min, max float64) bool {
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return num >= min && num <= max
}

// HasMinLength checks if a string has at least the minimum length.
func HasMinLength(s string, minLen int) bool {
	return len(s) >= minLen
}

// HasMaxLength checks if a string does not exceed the maximum length.
func HasMaxLength(s string, maxLen int) bool {
	return len(s) <= maxLen
}

// IsBase64 validates if a string is valid base64 encoding.
func IsBase64(s string) bool {
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	return len(s)%4 == 0 && base64Regex.MatchString(s)
}
