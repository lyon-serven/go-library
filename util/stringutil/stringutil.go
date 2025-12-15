// Package stringutil provides utility functions for string manipulation.
package stringutil

import (
	"strings"
)

// IsEmpty checks if a string is empty or contains only whitespace.
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Reverse returns the string with characters in reverse order.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Contains checks if a string contains a substring (case-sensitive).
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ContainsIgnoreCase checks if a string contains a substring (case-insensitive).
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Capitalize returns the string with the first letter capitalized.
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// TrimAndLower trims whitespace and converts to lowercase.
func TrimAndLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// SplitAndTrim splits a string by delimiter and trims each part.
func SplitAndTrim(s, delimiter string) []string {
	parts := strings.Split(s, delimiter)
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}

// PadLeft pads a string to the left with a specified character to reach the target length.
func PadLeft(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-len(s))
	return padding + s
}

// PadRight pads a string to the right with a specified character to reach the target length.
func PadRight(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-len(s))
	return s + padding
}

// RemoveSpaces removes all spaces from a string.
func RemoveSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
