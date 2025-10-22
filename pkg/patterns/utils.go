package patterns

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
