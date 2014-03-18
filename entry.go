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

func NewEntry(s string) Entry {
	if len(s) > 2 && s[0] == '<' && s[len(s)-1] == '>' {
		return newMatchEntry(s)
	}
	return newStaticEntry(s)
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

type BaseEntry struct {
	pattern  string
	handlers map[string]http.Handler
	entries  []Entry
}

func (e *BaseEntry) SetHandler(method string, h http.Handler) error {
	if e.HasHandler(method) {
		return errors.New("Duplicate Handler registration")
	}
	e.handlers[method] = h
	return nil
}

func (e *BaseEntry) HasHandler(method string) bool {
	_, ok := e.handlers[method]
	return ok
}

func (e *BaseEntry) Pattern() string {
	return e.pattern
}

func newStaticEntry(pattern string) *StaticEntry {
	base := BaseEntry{
		pattern,
		make(map[string]http.Handler),
		make([]Entry, 0),
	}
	return &StaticEntry{base}
}

type StaticEntry struct {
	BaseEntry
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
	BaseEntry
	name    string
	matcher Matcher
}

func newMatchEntry(pat string) *MatchEntry {
	s := pat[1 : len(pat)-1]
	ss := strings.Split(s, ":")
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
		panic(errors.New("no such match type: " + matchType))
	}

	e := MatchEntry{
		BaseEntry{pat, make(map[string]http.Handler), make([]Entry, 0)},
		name,
		matcher,
	}

	return &e
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
// 	BaseEntry
// 	name     string
// 	matcher  *SuffixMatcher
// }

// func newSuffixMatchEntry(pat string) *SuffixMatchEntry {

// 	return &SuffixMatchEntry{
// 		BaseEntry{pat, make(map[string]http.Handler), make([]Entry, 0)},
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
