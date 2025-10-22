package patterns

import (
	"strings"
	"unicode"
)

const AlphanumericClass = `\w`

// ContainsAlphanumericClass checks if a given string contains the alphanumeric class.
// This class corresponds to the Unicode category "Letter" and "Number", as well as the underscore character.
// The function returns true if the string contains the alphanumeric class, false otherwise.
func ContainsAlphanumericClass(input string) bool {
	return strings.Contains(input, AlphanumericClass)
}

// ContainsAlphanumeric checks if a given string contains alphanumeric characters.
// It returns true if the string contains alphanumeric characters, false otherwise.
// Alphanumeric characters are defined as Unicode letters (category "L"), Unicode
// digits (category "Nd"), and the underscore character.
func ContainsAlphanumeric(input []rune) bool {
	for _, r := range string(input) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return true
		}
	}
	return false
}
