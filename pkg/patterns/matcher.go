package patterns

import (
	"slices"
	"unicode"
)

// Pattern represents a sequence of pattern elements to match against
type Pattern struct {
	elements    []PatternElement
	startAnchor bool // true if pattern starts with ^
	endAnchor   bool // true if pattern ends with $
}

// PatternElement represents a single element in a pattern that can match runes
type PatternElement interface {
	Match(r rune) bool
}

// LiteralMatcher matches a specific rune
type LiteralMatcher struct {
	char rune
}

func (m LiteralMatcher) Match(r rune) bool {
	return m.char == r
}

// DigitMatcher matches any digit character
type DigitMatcher struct{}

func (m DigitMatcher) Match(r rune) bool {
	return unicode.IsDigit(r)
}

// AlphanumericMatcher matches any word character (letter, digit, or underscore)
type AlphanumericMatcher struct{}

func (m AlphanumericMatcher) Match(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// CharacterSetMatcher matches any character in a set of runes
type CharacterSetMatcher struct {
	chars   []rune
	negated bool
}

func (m CharacterSetMatcher) Match(r rune) bool {
	if slices.Contains(m.chars, r) {
		return !m.negated
	}
	return m.negated
}

// ParsePattern converts a pattern string into a sequence of pattern elements
func ParsePattern(pattern string) (*Pattern, error) {
	var elements []PatternElement
	runes := []rune(pattern)
	startAnchor := false
	endAnchor := false

	// Check for start anchor
	if len(runes) > 0 && runes[0] == '^' {
		startAnchor = true
		runes = runes[1:] // Remove the anchor from the pattern
	}

	// Check for end anchor
	if len(runes) > 0 && runes[len(runes)-1] == '$' {
		endAnchor = true
		runes = runes[:len(runes)-1] // Remove the end anchor
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch r {
		case '\\':
			if i+1 >= len(runes) {
				continue
			}
			i++
			switch runes[i] {
			case 'd':
				elements = append(elements, DigitMatcher{})
			case 'w':
				elements = append(elements, AlphanumericMatcher{})
			default:
				elements = append(elements, LiteralMatcher{char: runes[i]})
			}
		case '[':
			var chars []rune
			var negated bool
			i++
			if i < len(runes) && runes[i] == '^' {
				negated = true
				i++
			}
			for i < len(runes) && runes[i] != ']' {
				chars = append(chars, runes[i])
				i++
			}
			elements = append(elements, CharacterSetMatcher{chars: chars, negated: negated})
		default:
			elements = append(elements, LiteralMatcher{char: r})
		}
	}

	return &Pattern{elements: elements, startAnchor: startAnchor, endAnchor: endAnchor}, nil
}

// Match checks if a sequence of runes matches the pattern at any position
func (p *Pattern) Match(input []rune) bool {
	if p.startAnchor {
		// If pattern starts with ^, only try matching at the beginning
		return p.matchHere(input, 0)
	}

	// Otherwise, try matching at each position in the input
	for startPos := 0; startPos <= len(input); startPos++ {
		if p.matchHere(input, startPos) {
			return true
		}
	}
	return false
}

// matchHere attempts to match the pattern starting at the given position
func (p *Pattern) matchHere(input []rune, pos int) bool {
	patternPos := 0
	inputPos := pos

	// Try to match each pattern element in sequence
	for patternPos < len(p.elements) {
		if inputPos >= len(input) {
			return false
		}

		element := p.elements[patternPos]
		if !element.Match(input[inputPos]) {
			return false
		}

		patternPos++
		inputPos++
	}

	// If we have an end anchor, ensure we've reached the end of the input
	if p.endAnchor {
		return inputPos == len(input)
	}

	return true
}
