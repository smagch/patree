package treemux

import (
	"net/http"
	"strings"
)

type Entry interface {
	Exec(method, s string) (http.Handler, []string)
}

type StaticEntry struct {
	pattern  string
	handlers map[string]http.Handler
	entries  []Entry
}

func newStatic(p string) *StaticEntry {
	return &StaticEntry{
		pattern:  p,
		handlers: make(map[string]http.Handler),
		entries:  make([]Entry, 0),
	}
}

func (e *StaticEntry) Exec(method, str string) (http.Handler, []string) {
	if !strings.HasPrefix(str, e.pattern) {
		return nil, nil
	}

	var h http.Handler

	if len(str) == len(e.pattern) {
		h, _ = e.handlers[method]
		// TODO
		// create empty one? if handler exists???
		return h, nil
	}

	// TODO
	suffix := str[len(e.pattern):]
	for _, entry := range e.entries {
		var params []string
		if h, params = entry.Exec(method, suffix); h != nil {
			return h, params
		}
	}

	return nil, nil
}

func (e *StaticEntry) add(child Entry) {
	e.entries = append(e.entries, child)
}

type MatchEntry struct {
	name     string
	handlers map[string]http.Handler
	matcher  Matcher
	entries  []Entry
}

func newMatchEntry(n string, m Matcher) *MatchEntry {
	return &MatchEntry{
		name:     n,
		handlers: make(map[string]http.Handler),
		matcher:  m,
		entries:  make([]Entry, 0),
	}
}

func (e *MatchEntry) Exec(method, str string) (http.Handler, []string) {
	i := e.matcher.Match(str)
	if i == -1 {
		return nil, nil
	}

	// finish parsing
	if len(str) == i {
		h, _ := e.handlers[method]
		return h, []string{e.name, str}
	}

	for _, entry := range e.entries {
		if h, params := entry.Exec(method, str[i:]); h != nil {
			if params == nil {
				params = []string{e.name, str[:i]}
			} else {
				params = append(params, e.name, str[:i])
			}
			return h, params
		}
	}

	return nil, nil
}

type SuffixMatchEntry struct {
	name     string
	handlers map[string]http.Handler
	matcher  *SuffixMatcher
	entries  []Entry
}

func newSuffixMatchEntry(s string, m *SuffixMatcher) *SuffixMatchEntry {
	return &SuffixMatchEntry{
		name:     s,
		handlers: make(map[string]http.Handler),
		matcher:  m,
		entries:  make([]Entry, 0),
	}
}

func (e *SuffixMatchEntry) Exec(method, s string) (http.Handler, []string) {
	i := e.matcher.Match(s)
	if i == -1 {
		return nil, nil
	}

	// finish parsing
	if len(s) == i {
		h, _ := e.handlers[method]
		return h, []string{e.name, s[:(i - len(e.matcher.suffix))]}
	}

	for _, entry := range e.entries {
		if h, params := entry.Exec(method, s[i:]); h != nil {
			if params == nil {
				params = []string{e.name, s[:(i - len(e.matcher.suffix))]}
			} else {
				params = append(params, e.name, s[:(i-len(e.matcher.suffix))])
			}
			return h, params
		}
	}

	return nil, nil
}
