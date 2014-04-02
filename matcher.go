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
	Match(str string) (offset int, matchStr string)
	MatchRune(r rune) bool
}

type RuneMatcherFunc func(r rune) bool

func (f RuneMatcherFunc) Match(str string) (offset int, matchStr string) {
	offset = -1
	length := len(str)
	if length == 0 {
		return
	}

	for i, r := range str {
		if f(r) {
			continue
		}
		if i != 0 {
			offset = i
			matchStr = str[:i]
			return
		}
		return
	}

	offset = length
	matchStr = str
	return
}

func (f RuneMatcherFunc) MatchRune(r rune) bool {
	return f(r)
}

type SuffixMatcher struct {
	suffix  string
	matcher Matcher
}

func (m *SuffixMatcher) Match(str string) (offset int, matchStr string) {
	offset = -1

	// at least 1 character is required to match suffix and matcher
	d := len(str) - len(m.suffix)
	if d < 1 {
		return
	}

	for i, r := range str {
		if i > d {
			return
		}

		// peek string to match to suffix pattern
		if i != 0 && m.suffix == str[i:i+len(m.suffix)] {
			offset = i + len(m.suffix)
			matchStr = str[:i]
			return
		}

		if !m.matcher.MatchRune(r) {
			return
		}
	}
	return
}

func (m *SuffixMatcher) MatchRune(r rune) bool {
	return m.matcher.MatchRune(r)
}

var IntMatcher = RuneMatcherFunc(isDigit)
var HexMatcher = RuneMatcherFunc(isHex)
var DefaultMatcher = RuneMatcherFunc(isNotSlash)
