package patree

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

var NoClosingBracket = errors.New("Invalid syntax: No closing bracket found")

func isMatchPattern(s string) bool {
	return len(s) > 2 && s[0] == '<' && s[len(s)-1] == '>'
}

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
	if slashIndex > matchIndex && matchIndex != -1 {
		return matchIndex, data[:matchIndex], nil
	}

	// split by '/'
	// return data before '/' including '/'
	return slashIndex + 1, data[:(slashIndex + 1)], nil
}

// SplitPath splits the url pattern to entries.
func SplitPath(pat string) (routes []string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(pat))
	scanner.Split(routeSplitFunc)
	for scanner.Scan() {
		routes = append(routes, scanner.Text())
	}
	err = scanner.Err()
	return
}

func isNextSuffixPattern(p []string) bool {
	if len(p) >= 2 && isMatchPattern(p[0]) && !isMatchPattern(p[1]) {
		matcher, _ := parseMatcher(p[0])
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
