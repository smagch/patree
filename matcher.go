package treemux

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isHex(c byte) bool {
	return isDigit(c) || ('a' <= c && 'f' >= c)
}

type Matcher interface {
	Match(s string) int
}

type MatcherFunc func(s string) int

func (f MatcherFunc) Match(s string) int {
	return f(s)
}

type ByteMatcherFunc func(c byte) bool

func (f ByteMatcherFunc) Match(s string) int {
	length := len(s)
	if length == 0 || !f(s[0]) {
		return -1
	}
	for i := 1; i < length; i++ {
		if f(s[i]) {
			continue
		}
		return i
	}
	return length
}

var IntMatcher = ByteMatcherFunc(isDigit)
var HexMatcher = ByteMatcherFunc(isHex)
