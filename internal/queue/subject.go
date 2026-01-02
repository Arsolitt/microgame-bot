package queue

import "strings"

// matchSubject checks if subject matches the pattern with NATS-style wildcards:
// - '*' matches exactly one token
// - '>' matches one or more tokens (only at the end).
func matchSubject(subject, pattern string) bool {
	if pattern == subject {
		return true
	}

	subjectTokens := strings.Split(subject, ".")
	patternTokens := strings.Split(pattern, ".")

	si := 0
	pi := 0

	for pi < len(patternTokens) && si < len(subjectTokens) {
		switch patternTokens[pi] {
		case ">":
			// '>' must be the last token in pattern
			if pi != len(patternTokens)-1 {
				return false
			}
			return true
		case "*":
			// '*' matches exactly one token
			si++
			pi++
		default:
			// exact match required
			if patternTokens[pi] != subjectTokens[si] {
				return false
			}
			si++
			pi++
		}
	}

	// both must be fully consumed
	return si == len(subjectTokens) && pi == len(patternTokens)
}
