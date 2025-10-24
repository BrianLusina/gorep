package patterns

import (
	"fmt"
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

// OneOrMoreMatcher matches the underlying pattern one or more times
type OneOrMoreMatcher struct {
	matcher PatternElement
}

func (m OneOrMoreMatcher) Match(r rune) bool {
	return m.matcher.Match(r)
}

// ZeroOrOneMatcher matches the underlying pattern zero or one time
type ZeroOrOneMatcher struct {
	matcher PatternElement
}

func (m ZeroOrOneMatcher) Match(r rune) bool {
	return m.matcher.Match(r)
}

// WildcardMatcher matches any single character except newline
type WildcardMatcher struct{}

func (m WildcardMatcher) Match(r rune) bool {
	return r != '\n'
}

// AlternationMatcher matches one of several alternative patterns
type AlternationMatcher struct {
	alternatives []*Pattern
}

func (m AlternationMatcher) Match(r rune) bool {
	// This won't actually be called since we handle alternation specially in matchHere
	return false
}

// parseAlternatives parses a string containing alternatives separated by |
func parseAlternatives(pattern string) ([]*Pattern, error) {
	var alternatives []*Pattern
	var depth int

	// Split the pattern into alternatives, but only at top level
	runes := []rune(pattern)
	start := 0

	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '(':
			depth++
		case ')':
			depth--
		case '|':
			if depth == 0 {
				// Found a top-level alternation
				alt, err := ParsePattern(string(runes[start:i]))
				if err != nil {
					return nil, err
				}
				alternatives = append(alternatives, alt)
				start = i + 1
			}
		}
	}

	// Add the final alternative
	if start < len(runes) {
		alt, err := ParsePattern(string(runes[start:]))
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, alt)
	}

	return alternatives, nil
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
		case '(':
			// Find matching closing parenthesis
			start := i + 1
			depth := 1
			for i++; i < len(runes) && depth > 0; i++ {
				if runes[i] == '(' {
					depth++
				} else if runes[i] == ')' {
					depth--
				}
			}
			i-- // Step back to the closing parenthesis

			if depth > 0 {
				// Mismatched parentheses
				return nil, fmt.Errorf("missing closing parenthesis")
			}

			// Parse the alternatives within the parentheses
			innerPattern := string(runes[start:i])
			alternatives, err := parseAlternatives(innerPattern)
			if err != nil {
				return nil, err
			}

			element := AlternationMatcher{alternatives: alternatives}

			// Check for quantifiers after the parentheses
			if i+1 < len(runes) {
				switch runes[i+1] {
				case '+':
					i++
					elements = append(elements, OneOrMoreMatcher{matcher: element})
				case '?':
					i++
					elements = append(elements, ZeroOrOneMatcher{matcher: element})
				default:
					elements = append(elements, element)
				}
			} else {
				elements = append(elements, element)
			}

		case '.':
			element := WildcardMatcher{}
			// Check for quantifiers
			if i+1 < len(runes) {
				switch runes[i+1] {
				case '+':
					i++
					elements = append(elements, OneOrMoreMatcher{matcher: element})
				case '?':
					i++
					elements = append(elements, ZeroOrOneMatcher{matcher: element})
				default:
					elements = append(elements, element)
				}
			} else {
				elements = append(elements, element)
			}
		case '\\':
			if i+1 >= len(runes) {
				continue
			}
			i++
			var element PatternElement
			switch runes[i] {
			case 'd':
				element = DigitMatcher{}
			case 'w':
				element = AlphanumericMatcher{}
			default:
				element = LiteralMatcher{char: runes[i]}
			}
			// Check for quantifiers
			if i+1 < len(runes) {
				switch runes[i+1] {
				case '+':
					i++
					elements = append(elements, OneOrMoreMatcher{matcher: element})
				case '?':
					i++
					elements = append(elements, ZeroOrOneMatcher{matcher: element})
				default:
					elements = append(elements, element)
				}
			} else {
				elements = append(elements, element)
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
			element := CharacterSetMatcher{chars: chars, negated: negated}
			// Check for quantifiers after the character set
			if i+1 < len(runes) {
				switch runes[i+1] {
				case '+':
					i++
					elements = append(elements, OneOrMoreMatcher{matcher: element})
				case '?':
					i++
					elements = append(elements, ZeroOrOneMatcher{matcher: element})
				default:
					elements = append(elements, element)
				}
			} else {
				elements = append(elements, element)
			}
		default:
			element := LiteralMatcher{char: r}
			// Check for quantifiers
			if i+1 < len(runes) {
				switch runes[i+1] {
				case '+':
					i++
					elements = append(elements, OneOrMoreMatcher{matcher: element})
				case '?':
					i++
					elements = append(elements, ZeroOrOneMatcher{matcher: element})
				default:
					elements = append(elements, element)
				}
			} else {
				elements = append(elements, element)
			}
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
		element := p.elements[patternPos]

		// Handle quantifiers
		switch q := element.(type) {
		case AlternationMatcher:
			// Try each alternative from the current position
			remainingPattern := &Pattern{
				elements:  p.elements[patternPos+1:],
				endAnchor: p.endAnchor,
			}

			for _, alt := range q.alternatives {
				// Create a pattern that combines this alternative with the remaining elements
				altPattern := &Pattern{
					elements:  append(alt.elements, remainingPattern.elements...),
					endAnchor: p.endAnchor,
				}
				if altPattern.matchHere(input, inputPos) {
					return true
				}
			}
			return false

		case OneOrMoreMatcher:
			// Must match at least once, so we need input
			if inputPos >= len(input) {
				return false
			}
			if !q.Match(input[inputPos]) {
				return false
			}

			inputPos++
			// Match additional occurrences
			for inputPos < len(input) && q.Match(input[inputPos]) {
				inputPos++
			}

			// After matching one or more, try to match the rest of the pattern
			// at any position from here onwards
			remainingPattern := &Pattern{
				elements:  p.elements[patternPos+1:],
				endAnchor: p.endAnchor,
			}

			// Try each possible position after our matches
			for tryPos := inputPos; tryPos >= pos+1; tryPos-- {
				if remainingPattern.matchHere(input, tryPos) {
					return true
				}
			}
			return false

		case ZeroOrOneMatcher:
			// Try skipping the optional element first (zero case)
			remainingPattern := &Pattern{
				elements:  p.elements[patternPos+1:],
				endAnchor: p.endAnchor,
			}
			if remainingPattern.matchHere(input, inputPos) {
				return true
			}

			// If we still have input, try matching the element once
			if inputPos < len(input) && q.Match(input[inputPos]) {
				if remainingPattern.matchHere(input, inputPos+1) {
					return true
				}
			}

			return false

		default:
			// Normal (non-quantified) element needs input to match
			if inputPos >= len(input) {
				return false
			}
			if !element.Match(input[inputPos]) {
				return false
			}
			patternPos++
			inputPos++
			continue
		}
	}

	// If we have an end anchor, ensure we've reached the end of the input
	if p.endAnchor {
		return inputPos == len(input)
	}

	return true
}
