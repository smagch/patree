package treemux

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var matcherMap = map[string]Matcher{
	"default": DefaultMatcher,
	"int":     IntMatcher,
	"hex":     HexMatcher,
}

func NewEntry(s string) (Entry, error) {
	// return static entry if the first char isn't '<'
	if s[0] != '<' {
		return newStaticEntry(s), nil
	}
	return newMatchEntry(s)
}

//
// Entry
//
type Entry interface {
	Exec(method, s string) (http.Handler, []string)
	Pattern() string
	HasHandler(method string) bool
	SetHandler(method string, h http.Handler) error
}

type StaticEntry struct {
	pattern  string
	handlers map[string]http.Handler
	entries  []Entry
}

func newStaticEntry(pattern string) *StaticEntry {
	return &StaticEntry{
		pattern,
		make(map[string]http.Handler),
		make([]Entry, 0),
	}
}

func (p *StaticEntry) SetHandler(method string, h http.Handler) error {
	if p.HasHandler(method) {
		return errors.New("Duplicate Handler registration")
	}
	p.handlers[method] = h
	return nil
}

func (p *StaticEntry) HasHandler(method string) bool {
	_, ok := p.handlers[method]
	return ok
}

func (p *StaticEntry) Pattern() string {
	return p.pattern
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

//
// Match
//
type MatchEntry struct {
	StaticEntry
	name    string
	matcher Matcher
}

func newMatchEntry(pat string) (*MatchEntry, error) {
	if pat[0] != '<' || pat[len(pat)-1] != '>' {
		return nil, errors.New("invalid pattern: " + pat)
	}

	s := pat[1 : len(pat)-1]
	ss := strings.Split(s, ":")
	if len(ss) > 2 {
		return nil, fmt.Errorf("invalid match syntax: %s. Only one ':' is allowed", s)
	}

	var name, matchType string
	if len(ss) == 1 {
		name = ss[0]
	} else {
		matchType = ss[0]
		name = ss[1]
	}
	if matchType == "" {
		matchType = "default"
	}

	matcher := matcherMap[matchType]
	if matcher == nil {
		return nil, errors.New("no such match type: " + matchType)
	}

	e := MatchEntry{
		StaticEntry{pat, make(map[string]http.Handler), make([]Entry, 0)},
		name,
		matcher,
	}

	return &e, nil
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

//
// Suffix
//
// type SuffixMatchEntry struct {
// 	StaticEntry
// 	name     string
// 	matcher  *SuffixMatcher
// }

// func newSuffixMatchEntry(pat string) *SuffixMatchEntry {

// 	return &SuffixMatchEntry{
// 		StaticEntry{pat, make(map[string]http.Handler), make([]Entry, 0)},
// 		pat,
// 		matcher: m,
// 	}
// }

// func (e *SuffixMatchEntry) Exec(method, s string) (http.Handler, []string) {
// 	i := e.matcher.Match(s)
// 	if i == -1 {
// 		return nil, nil
// 	}

// 	// finish parsing
// 	if len(s) == i {
// 		h, _ := e.handlers[method]
// 		return h, []string{e.name, s[:(i - len(e.matcher.suffix))]}
// 	}

// 	for _, entry := range e.entries {
// 		if h, params := entry.Exec(method, s[i:]); h != nil {
// 			if params == nil {
// 				params = []string{e.name, s[:(i - len(e.matcher.suffix))]}
// 			} else {
// 				params = append(params, e.name, s[:(i-len(e.matcher.suffix))])
// 			}
// 			return h, params
// 		}
// 	}

// 	return nil, nil
// }
