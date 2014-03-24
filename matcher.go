package patree

import (
	"errors"
	"strings"
)

func parseMatcher(pat string) (matcher Matcher, name string) {
	s := pat[1 : len(pat)-1]
	ss := strings.Split(s, ":")
	var matchType string
	if len(ss) == 1 {
		name = ss[0]
	} else {
		matchType = ss[0]
		name = ss[1]
	}
	if matchType == "" {
		matchType = "default"
	}

	matcher = matcherMap[matchType]
	if matcher == nil {
		panic(errors.New("no such match type: " + matchType))
	}

	return matcher, name
}

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
	MatchRune(r rune) bool
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

func (f RuneMatcherFunc) MatchRune(r rune) bool {
	return f(r)
}

type SuffixMatcher struct {
	suffix  string
	matcher Matcher
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

		if !m.matcher.MatchRune(r) {
			return -1
		}
	}

	return -1
}

func (m *SuffixMatcher) MatchRune(r rune) bool {
	return m.matcher.MatchRune(r)
}

var IntMatcher = RuneMatcherFunc(isDigit)
var HexMatcher = RuneMatcherFunc(isHex)
var DefaultMatcher = RuneMatcherFunc(isNotSlash)
