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

type IntMatcher struct {}

func (m *IntMatcher) Match(str string) int {
	length := len(str)
	if length == 0 || !isDigit(str[0]) {
		return -1
	}
	for i := 1; i < length; i++ {
		if isDigit(str[i]) {
			continue
		}
		return i
	}
	return length
}
