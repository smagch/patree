package patree

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

// NoClosingBracket is the error returned by SplitPath when there is no closing
// bracket '>' after opening bracket '<'.
var NoClosingBracket = errors.New("Invalid syntax: No closing bracket found")

// MatherMap stores Matchers with matcher pattern type keys. For example,
// Pattern "<int:id>" is an IntMatcher.
// Pattern "<hex:id> is a HexMatcher.
// Pattern "<id>" is a DefaultMatcher.
// Pattern "<uuid:id>" is a UUIDMatcher
var MatcherMap = map[string]Matcher{
	"default": DefaultMatcher,
	"int":     IntMatcher,
	"hex":     HexMatcher,
	"uuid":    UUIDMatcher,
}

// parseMatcher returns matcher and name from the given pattern string.
func parseMatcher(pat string) (matcher Matcher, name string) {
	if !isMatchPattern(pat) {
		panic("pattern \"" + pat + "\" is not a matcher pattern")
	}

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

	matcher = MatcherMap[matchType]
	if matcher == nil {
		panic(errors.New("no such match type: " + matchType))
	}

	return matcher, name
}

// isMatchPattern see if given string is match pattern.
func isMatchPattern(s string) bool {
	return len(s) > 2 && s[0] == '<' && s[len(s)-1] == '>'
}

// routeSplitFunc is the SplitFunc to scan url pattern.
func routeSplitFunc(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF || data == nil {
		return 0, nil, io.EOF
	}

	r, _ := utf8.DecodeRune(data)

	// matcher
	if r == '<' {
		i := bytes.IndexRune(data, '>')
		if i == -1 {
			return 0, nil, NoClosingBracket
		}
		return (i + 1), data[:(i + 1)], nil
	}

	// one char
	if len(data) == 1 {
		return 1, data, nil
	}

	// should ignore first '/'
	slashIndex := bytes.IndexRune(data[1:], '/')
	if slashIndex != -1 {
		slashIndex++
	}

	matchIndex := bytes.IndexRune(data, '<')

	// remaining string would be a static entry
	if slashIndex == -1 && matchIndex == -1 {
		return len(data), data, nil
	}

	// split by '<'
	// return data before '<'
	if matchIndex != -1 && (slashIndex == -1 || slashIndex > matchIndex) {
		return matchIndex, data[:matchIndex], nil
	}

	// split by '/'
	// return data before '/' including '/'
	return slashIndex + 1, data[:(slashIndex + 1)], nil
}

// SplitPath splits the url pattern.
func SplitPath(pat string) (routes []string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(pat))
	scanner.Split(routeSplitFunc)
	for scanner.Scan() {
		routes = append(routes, scanner.Text())
	}
	err = scanner.Err()
	return
}

// isNextSuffixPattern see next 2 patterns can be suffix matcher. If following
// two cases are both true, it could possibly become a suffix matcher.
//   1. the first pattern is a Matcher.
//   2. the next pattern is a static pattern.
// If the first matcher can't match the first rune of the second static pattern,
// next pattern should be a suffix matcher combined the two patterns.
func isNextSuffixPattern(p []string) bool {
	if len(p) >= 2 && isMatchPattern(p[0]) && !isMatchPattern(p[1]) {
		matcher, _ := parseMatcher(p[0])
		if _, ok := matcher.(*FixedLengthMatcher); ok {
			return false
		}
		r, _ := utf8.DecodeRuneInString(p[1])
		return matcher.MatchRune(r)
	}
	return false
}

// PeekNextPattern returns next entry pattern with offset size
func PeekNextPattern(p []string) (pat string, size int) {
	if isNextSuffixPattern(p) {
		pat, size = (p[0] + p[1]), 2
	} else {
		pat, size = p[0], 1
	}
	return
}
