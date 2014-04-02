package patree

import (
	"errors"
	"net/http"
	"strings"
)

var matcherMap = map[string]Matcher{
	"default": DefaultMatcher,
	"int":     IntMatcher,
	"hex":     HexMatcher,
}

// ExecFunc is pattern match function that returns handler and matched request
// url parameters.
type ExecFunc func(method, urlStr string) (http.Handler, []string)

// Entry is a pattern node.
type Entry struct {
	pattern  string
	handlers map[string]http.Handler
	handler  http.Handler
	entries  []*Entry
	exec     ExecFunc
}

// Len returns a total number of child entries.
func (e *Entry) Len() int {
	return len(e.entries)
}

// SetHandler reigsters the given handler that matches with any method.
func (e *Entry) SetHandler(h http.Handler) error {
	if e.handler != nil {
		return errors.New("Duplicate Handler registration")
	}
	e.handler = h
	return nil
}

// SetMethodHandler reigsters the given handler for the method.
func (e *Entry) SetMethodHandler(method string, h http.Handler) error {
	if e.GetHandler(method) != nil {
		return errors.New("Duplicate Handler registration")
	}
	e.handlers[method] = h
	return nil
}

// GetHandler returns a handler with given method.
func (e *Entry) GetHandler(method string) http.Handler {
	handler := e.handlers[method]
	if handler == nil {
		handler = e.handler
	}
	return handler
}

// Pattern returns a string that the entry represents.
func (e *Entry) Pattern() string {
	return e.pattern
}

// getChildEntry returns a child Entry that matches the given pattern string.
func (e *Entry) getChildEntry(pat string) *Entry {
	for _, entry := range e.entries {
		if pat == entry.Pattern() {
			return entry
		}
	}
	return nil
}

// MergePattern add entry patterns with given pattern strings. If a pattern
// already exists on the entry, it adds remaining patterns to the existing entry.
func (e *Entry) MergePatterns(patterns []string) *Entry {
	pat, size := PeekNextPattern(patterns)
	if child := e.getChildEntry(pat); child != nil {
		if len(patterns) == size {
			return child
		}
		return child.MergePatterns(patterns[size:])
	}
	return e.addPatterns(patterns)
}

// AddEntry add new child entry.
//
// TODO sort
func (e *Entry) AddEntry(child *Entry) {
	e.entries = append(e.entries, child)
}

// addPatterns adds entry children with the pattern strings.
func (e *Entry) addPatterns(patterns []string) *Entry {
	var currentNode *Entry = e
	for len(patterns) > 0 {
		var entry *Entry
		pat, size := PeekNextPattern(patterns)

		// suffix entry
		if size == 2 {
			matcher, name := parseMatcher(patterns[0])
			suffixMatcher := &SuffixMatcher{patterns[1], matcher}
			entry = newSuffixMatchEntry(pat, name, suffixMatcher)
		} else if isMatchPattern(pat) {
			entry = newMatchEntry(pat)
		} else {
			entry = newStaticEntry(pat)
		}

		currentNode.AddEntry(entry)
		currentNode = entry
		patterns = patterns[size:]
	}
	return currentNode
}

func newEntry(pat string) *Entry {
	return &Entry{
		pattern:  pat,
		handlers: make(map[string]http.Handler),
	}
}

func newStaticEntry(pat string) *Entry {
	entry := newEntry(pat)
	entry.exec = entry.execPrefix
	return entry
}

func newMatchEntry(pat string) *Entry {
	entry := newEntry(pat)
	matcher, name := parseMatcher(pat)
	entry.exec = entry.getExecMatch(name, matcher)
	return entry
}

func newSuffixMatchEntry(pat, name string, matcher Matcher) *Entry {
	entry := newEntry(pat)
	entry.exec = entry.getExecMatch(name, matcher)
	return entry
}

// execPrefix simply see if the given urlStr has a leading pattern.
func (e *Entry) execPrefix(method, urlStr string) (http.Handler, []string) {
	if !strings.HasPrefix(urlStr, e.pattern) {
		return nil, nil
	}
	if len(urlStr) == len(e.pattern) {
		return e.GetHandler(method), nil
	}
	return e.traverse(method, urlStr[len(e.pattern):])
}

// traverse tries matches to child entries.
func (e *Entry) traverse(method, urlStr string) (http.Handler, []string) {
	for _, entry := range e.entries {
		if h, params := entry.exec(method, urlStr); h != nil {
			return h, params
		}
	}
	return nil, nil
}

// getExecMatch returns ExecFunc with the given name and mather.
func (e *Entry) getExecMatch(name string, matcher Matcher) ExecFunc {
	return func(method, urlStr string) (http.Handler, []string) {
		offset, matchStr := matcher.Match(urlStr)
		if offset == -1 {
			return nil, nil
		}

		// finish parsing
		if len(urlStr) == offset {
			if h := e.GetHandler(method); h != nil {
				return h, []string{name, matchStr}
			}
			return nil, nil
		}

		for _, entry := range e.entries {
			if h, params := entry.exec(method, urlStr[offset:]); h != nil {
				params = append(params, name, matchStr)
				return h, params
			}
		}
		return nil, nil
	}
}
