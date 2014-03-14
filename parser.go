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

func routeSplitFunc(data []byte, atEOF bool) (int, []byte, error) {
	// var b bytes.Buffer
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

	matchIndex := bytes.IndexRune(data, '<')
	slashIndex := bytes.IndexRune(data[1:], '/')
	if slashIndex != -1 {
		slashIndex++
	}
	if slashIndex == -1 && matchIndex == -1 {
		return len(data), data, nil
	}

	if slashIndex > matchIndex {
		return slashIndex + 1, data[:(slashIndex + 1)], nil
	}

	return matchIndex, data[:matchIndex], nil
}

func SplitPath(s string) []string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(routeSplitFunc)
	routes := make([]string, 0)
	for scanner.Scan() {
		routes = append(routes, scanner.Text())
	}
	return routes
}
