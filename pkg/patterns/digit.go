package patterns

import (
	"strings"
	"unicode"
)

const DigitClass = `\d`

func ContainsDigitClass(input string) bool {
	return strings.Contains(input, DigitClass)
}

func ContainsDigit(input []rune) bool {
	for _, r := range string(input) {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
