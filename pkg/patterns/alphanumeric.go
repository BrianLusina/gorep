package patterns

import (
	"strings"
	"unicode"
)

const AlphanumericClass = `\w`

func ContainsAlphanumericClass(input string) bool {
	return strings.Contains(input, AlphanumericClass)
}

func ContainsAlphanumeric(input []rune) bool {
	for _, r := range string(input) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return true
		}
	}
	return false
}
