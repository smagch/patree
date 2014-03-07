package treemux

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isHex(r rune) bool {
	return isDigit(r) || ('a' <= r && 'f' >= r)
}

func isNotSlash(r rune) bool {
	return r != '/'
}

type Matcher interface {
	Match(s string) int
}

type MatcherFunc func(s string) int

func (f MatcherFunc) Match(s string) int {
	return f(s)
}

type RuneMatcherFunc func(r rune) bool

func (f RuneMatcherFunc) Match(s string) int {
	length := len(s)
	if length == 0 {
		return -1
	}

	for i, r := range s {
		if f(r) {
			continue
		}
		if i != 0 {
			return i
		}
		return -1
	}

	return length
}

type SuffixMatcher struct {
	suffix string
	f      RuneMatcherFunc
}

func (m *SuffixMatcher) Match(s string) int {
	d := len(s) - len(m.suffix)
	// at least 1 character is required to match suffix and matcher
	if d < 1 {
		return -1
	}

	for i, r := range s {
		if i > d {
			return -1
		}

		// peek string to match to suffix pattern
		if i != 0 && m.suffix == s[i:i+len(m.suffix)] {
			return i + len(m.suffix)
		}

		if !m.f(r) {
			return -1
		}
	}

	return -1
}

var IntMatcher = RuneMatcherFunc(isDigit)
var HexMatcher = RuneMatcherFunc(isHex)
var DefaultMatcher = RuneMatcherFunc(isNotSlash)
