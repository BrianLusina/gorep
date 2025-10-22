package patterns

import (
	"strings"
	"unicode"
)

const DigitClass = `\d`

// ContainsDigitClass checks if a given string contains the digit class.
// This class corresponds to the Unicode category "Nd", which is composed of
// all Unicode digits.
// The function returns true if the string contains the digit class, false otherwise.
func ContainsDigitClass(input string) bool {
	return strings.Contains(input, DigitClass)
}

// ContainsDigit checks if a given string contains Unicode digits.
// It returns true if the string contains Unicode digits, false otherwise.
// Unicode digits are defined as characters in the Unicode category "Nd".
func ContainsDigit(input []rune) bool {
	for _, r := range string(input) {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
