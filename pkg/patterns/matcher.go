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
	groupCount  int  // number of capturing groups in the pattern
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

// GroupMatcher represents a capturing group; index is 1-based
type GroupMatcher struct {
	index   int
	pattern *Pattern
}

func (m GroupMatcher) Match(r rune) bool {
	// Not used directly; matching uses the group's pattern
	return false
}

// BackReferenceMatcher matches the previously captured group text
type BackReferenceMatcher struct {
	index int
}

func (m BackReferenceMatcher) Match(r rune) bool {
	// Not used directly; matching compares substrings
	return false
}

// parseAlternatives parses a string containing alternatives separated by |
func parseAlternatives(pattern string, groupCounter *int) ([]*Pattern, error) {
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
				alt, err := parsePatternInternal(string(runes[start:i]), groupCounter)
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
		alt, err := parsePatternInternal(string(runes[start:]), groupCounter)
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, alt)
	}

	return alternatives, nil
}

// ParsePattern converts a pattern string into a sequence of pattern elements
// parsePatternInternal parses a pattern and updates groupCounter for capturing groups
func parsePatternInternal(pattern string, groupCounter *int) (*Pattern, error) {
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
			// This is a capturing group: assign group index
			(*groupCounter)++
			groupIndex := *groupCounter

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

			// Parse the content within the parentheses (may include alternation)
			innerPatternStr := string(runes[start:i])
			var inner *Pattern
			// Parse alternatives inside group
			alts, err := parseAlternatives(innerPatternStr, groupCounter)
			if err != nil {
				return nil, err
			}
			if len(alts) == 1 {
				inner = alts[0]
			} else {
				// Build a pattern whose single element is an AlternationMatcher
				alternation := AlternationMatcher{alternatives: alts}
				inner = &Pattern{elements: []PatternElement{alternation}}
			}
			if inner != nil {
				inner.groupCount = *groupCounter
			}

			element := GroupMatcher{index: groupIndex, pattern: inner}

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
			// Backreference if digit follows
			if runes[i] >= '1' && runes[i] <= '9' {
				idx := int(runes[i] - '0')
				element := BackReferenceMatcher{index: idx}
				// quantifiers are not meaningful on backreferences here; just append
				elements = append(elements, element)
				continue
			}

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

	return &Pattern{elements: elements, startAnchor: startAnchor, endAnchor: endAnchor, groupCount: *groupCounter}, nil
}

// ParsePattern is the public entry that initializes group counting
func ParsePattern(pattern string) (*Pattern, error) {
	counter := 0
	p, err := parsePatternInternal(pattern, &counter)
	if err != nil {
		return nil, err
	}
	p.groupCount = counter
	return p, nil
}

// matchElementOnce attempts to match a single occurrence of element at pos.
// It returns (matched, newPos, updatedCaptures).
func matchElementOnce(element PatternElement, input []rune, pos int, captures []string, p *Pattern) (bool, int, []string) {
	switch e := element.(type) {
	case GroupMatcher:
		// Match the group's inner pattern starting at pos
		cp := make([]string, len(captures))
		copy(cp, captures)
		var matchEnd int
		// Try to match the inner pattern
		if ok, newPos := e.pattern.matchHereWithCaptures(input, pos, cp); ok {
			matchEnd = newPos
			// store the captured substring - use the actual matched range
			cp[e.index-1] = string(input[pos:matchEnd])
			return true, matchEnd, cp
		}
		return false, 0, nil
	case BackReferenceMatcher:
		if e.index-1 < 0 || e.index-1 >= len(captures) {
			return false, 0, nil
		}
		capStr := captures[e.index-1]
		if capStr == "" {
			// No capture yet for this group, can't match
			return false, 0, nil
		}
		capRunes := []rune(capStr)
		if pos+len(capRunes) > len(input) {
			// Not enough input remaining to match the captured text
			return false, 0, nil
		}
		// Compare the captured text with input at current position
		matched := true
		for i := 0; i < len(capRunes); i++ {
			if input[pos+i] != capRunes[i] {
				matched = false
				break
			}
		}
		if matched {
			// Return the captures unchanged since backreferences don't create new captures
			return true, pos + len(capRunes), captures
		}
		return false, 0, nil
	default:
		// Simple rune-based element
		if pos >= len(input) {
			return false, 0, nil
		}
		if !element.Match(input[pos]) {
			return false, 0, nil
		}
		return true, pos + 1, captures
	}
}

// Match checks if a sequence of runes matches the pattern at any position
func (p *Pattern) Match(input []rune) bool {
	if p.startAnchor {
		captures := make([]string, p.groupCount)
		ok, _, _ := p.matchHereWithState(input, 0, captures)
		return ok
	}

	// Try matching at each position
	for startPos := 0; startPos <= len(input); startPos++ {
		captures := make([]string, p.groupCount)
		if ok, _, _ := p.matchHereWithState(input, startPos, captures); ok {
			return true
		}
	}
	return false
}

// matchHereWithState attempts to match at current position and manages captured groups
func (p *Pattern) matchHereWithState(input []rune, pos int, captures []string) (bool, int, []string) {
	if len(p.elements) == 0 {
		// Empty pattern matches if no end anchor or if we're at end of input
		if !p.endAnchor || pos == len(input) {
			return true, pos, captures
		}
		return false, pos, captures
	}

	element := p.elements[0]
	remaining := &Pattern{
		elements:   p.elements[1:],
		endAnchor:  p.endAnchor,
		groupCount: p.groupCount,
	}

	switch e := element.(type) {
	case GroupMatcher:
		if ok, newPos, _ := e.pattern.matchHereWithState(input, pos, captures); ok {
			// Store the group's match
			captures[e.index-1] = string(input[pos:newPos])
			// Try the rest of the pattern
			if ok2, finalPos, finalCaptures := remaining.matchHereWithState(input, newPos, captures); ok2 {
				return true, finalPos, finalCaptures
			}
		}
		return false, pos, captures

	case BackReferenceMatcher:
		if e.index < 1 || e.index > len(captures) {
			return false, pos, captures
		}
		captured := captures[e.index-1]
		if captured == "" {
			return false, pos, captures
		}
		if pos+len(captured) > len(input) {
			return false, pos, captures
		}
		// Must match exactly what was captured before
		if string(input[pos:pos+len(captured)]) != captured {
			return false, pos, captures
		}
		// Try remaining pattern after the backreference
		return remaining.matchHereWithState(input, pos+len(captured), captures)

	case OneOrMoreMatcher:
		// Must match at least once
		if ok, newPos, newCaptures := matchElementOnce(e.matcher, input, pos, captures, p); !ok {
			return false, pos, captures
		} else {
			pos = newPos
			captures = newCaptures
		}

		// Try matching remaining pattern at each position after matching at least one
		for currentPos := pos; currentPos <= len(input); currentPos++ {
			// Try remaining pattern at current position
			if ok, finalPos, finalCaptures := remaining.matchHereWithState(input, currentPos, captures); ok {
				return true, finalPos, finalCaptures
			}

			// Try to match one more at current position
			if currentPos < len(input) && e.matcher.Match(input[currentPos]) {
				pos = currentPos + 1
			} else {
				break
			}
		}
		return false, pos, captures

	case ZeroOrOneMatcher:
		// Try skipping first
		if ok, newPos, newCaptures := remaining.matchHereWithState(input, pos, captures); ok {
			return true, newPos, newCaptures
		}
		// Try matching once
		if ok, newPos, newCaptures := matchElementOnce(e.matcher, input, pos, captures, p); ok {
			return remaining.matchHereWithState(input, newPos, newCaptures)
		}
		return false, pos, captures

	case AlternationMatcher:
		for _, alt := range e.alternatives {
			if ok, newPos, newCaptures := alt.matchHereWithState(input, pos, captures); ok {
				return remaining.matchHereWithState(input, newPos, newCaptures)
			}
		}
		return false, pos, captures

	default:
		if ok, newPos, newCaptures := matchElementOnce(element, input, pos, captures, p); ok {
			return remaining.matchHereWithState(input, newPos, newCaptures)
		}
		return false, pos, captures
	}
}

// matchHere attempts to match the pattern starting at the given position
// matchHereWithCaptures attempts to match the pattern starting at pos using captures.
// It returns (matched, newPos). captures is mutated on successful paths.
func (p *Pattern) matchHereWithCaptures(input []rune, pos int, captures []string) (bool, int) {
	patternPos := 0
	inputPos := pos

	for patternPos < len(p.elements) {
		element := p.elements[patternPos]

		switch q := element.(type) {
		case AlternationMatcher:
			remainingPattern := &Pattern{
				elements:   p.elements[patternPos+1:],
				endAnchor:  p.endAnchor,
				groupCount: p.groupCount,
			}

			for _, alt := range q.alternatives {
				// Try this alternative
				cp := make([]string, len(captures))
				copy(cp, captures)

				// First match the alternative
				if ok, altPos := alt.matchHereWithCaptures(input, inputPos, cp); ok {
					// Then try to match the remaining pattern
					if ok2, finalPos := remainingPattern.matchHereWithCaptures(input, altPos, cp); ok2 {
						copy(captures, cp)
						return true, finalPos
					}
				}
			}
			return false, 0

		case OneOrMoreMatcher:
			// Need at least one match
			cp := make([]string, len(captures))
			copy(cp, captures)
			ok, newPos, newCp := matchElementOnce(q.matcher, input, inputPos, cp, p)
			if !ok {
				return false, 0
			}
			inputPos = newPos
			copy(cp, newCp)

			// Try matching the remaining pattern after each repetition
			remainingPattern := &Pattern{
				elements:   p.elements[patternPos+1:],
				endAnchor:  p.endAnchor,
				groupCount: p.groupCount,
			}

			// Keep trying to consume more occurrences while attempting the remainder
			for {
				// Try the remaining pattern at current position
				tryCp := make([]string, len(cp))
				copy(tryCp, cp)
				if ok, newPosRem := remainingPattern.matchHereWithCaptures(input, inputPos, tryCp); ok {
					copy(captures, tryCp)
					return true, newPosRem
				}

				// Try one more occurrence
				ok2, nextPos, nextCp := matchElementOnce(q.matcher, input, inputPos, cp, p)
				if !ok2 {
					break
				}
				inputPos = nextPos
				copy(cp, nextCp)
			}
			return false, 0

		case ZeroOrOneMatcher:
			remainingPattern := &Pattern{
				elements:   p.elements[patternPos+1:],
				endAnchor:  p.endAnchor,
				groupCount: p.groupCount,
			}

			// First try zero occurrences (skip the element)
			cpZero := make([]string, len(captures))
			copy(cpZero, captures)
			if ok, newPos := remainingPattern.matchHereWithCaptures(input, inputPos, cpZero); ok {
				copy(captures, cpZero)
				return true, newPos
			}

			// Then try matching one occurrence
			cp := make([]string, len(captures))
			copy(cp, captures)
			ok, newPos, newCp := matchElementOnce(q.matcher, input, inputPos, cp, p)
			if ok {
				// Try to match the remainder after this occurrence
				if ok2, newPos2 := remainingPattern.matchHereWithCaptures(input, newPos, newCp); ok2 {
					copy(captures, newCp)
					return true, newPos2
				}
			}
			return false, 0

		default:
			// Normal (non-quantified) element: attempt to match once
			ok, newPos, newCp := matchElementOnce(element, input, inputPos, captures, p)
			if !ok {
				return false, 0
			}
			// update captures and advance
			copy(captures, newCp)
			inputPos = newPos
			patternPos++
			continue
		}
	}

	// If we have an end anchor, ensure we've reached the end of the input
	if p.endAnchor {
		return inputPos == len(input), inputPos
	}
	return true, inputPos
}
