package treemux

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isHex(r rune) bool {
	return isDigit(r) || ('a' <= r && 'f' >= r)
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

var IntMatcher = RuneMatcherFunc(isDigit)
var HexMatcher = RuneMatcherFunc(isHex)
