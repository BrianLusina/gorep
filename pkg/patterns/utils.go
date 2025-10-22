package patterns

// ParseGroups takes a pattern string and returns two slices of runes.
// The first slice contains the positive characters in the pattern and the second slice contains the negative characters.
// Positive characters are those that are inside the [] character block and are not preceded by a ^ character.
// Negative characters are those that are inside the [] character block and are preceded by a ^ character.
// Characters outside the [] character block are ignored.
func ParseGroups(pattern string) ([]rune, []rune) {
	var positive []rune
	var negative []rune
	add := false
	neg := false
	for _, r := range pattern {
		switch r {
		case '[':
			add = true
		case ']':
			add = false
			neg = false
		case '^':
			neg = true
		default:
			if add {
				if neg {
					negative = append(negative, r)
				} else {
					positive = append(positive, r)
				}
			} else {
				//
			}
		}
	}
	return positive, negative
}
