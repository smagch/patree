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
	AddEntry(child Entry)
	Len() int
	MergePatterns([]string) Entry
	SetHandler(h http.Handler) error
	SetMethodHandler(method string, h http.Handler) error
	GetHandler(method string) http.Handler
}

type StaticEntry struct {
	pattern  string
	handlers map[string]http.Handler
	handler  http.Handler
	entries  []Entry
}

func (e *StaticEntry) Len() int {
	return len(e.entries)
}

func (e *StaticEntry) SetHandler(h http.Handler) error {
	if e.handler != nil {
		return errors.New("Duplicate Handler registration")
	}
	e.handler = h
	return nil
}

func (e *StaticEntry) SetMethodHandler(method string, h http.Handler) error {
	if e.GetHandler(method) != nil {
		return errors.New("Duplicate Handler registration")
	}
	e.handlers[method] = h
	return nil
}

func (e *StaticEntry) GetHandler(method string) http.Handler {
	handler := e.handlers[method]
	if handler == nil {
		handler = e.handler
	}
	return handler
}

func (e *StaticEntry) Pattern() string {
	return e.pattern
}

func (e *StaticEntry) getChildEntry(pat string) Entry {
	for _, entry := range e.entries {
		if pat == entry.Pattern() {
			return entry
		}
	}
	return nil
}

func (e *StaticEntry) MergePatterns(patterns []string) Entry {
	pat := patterns[0]
	if child := e.getChildEntry(pat); child != nil {
		if len(patterns) == 1 {
			panic(errors.New("duplicate pattern: " + pat))
		}
		return child.MergePatterns(patterns[1:])
	}
	return e.addPatterns(patterns)
}

// TODO sort
func (e *StaticEntry) AddEntry(child Entry) {
	e.entries = append(e.entries, child)
}

func (e *StaticEntry) addPatterns(patterns []string) Entry {
	var currentNode Entry = Entry(e)
	for _, pat := range patterns {
		entry := NewEntry(pat)
		currentNode.AddEntry(entry)
		currentNode = entry
	}
	return currentNode
}

func newStaticEntry(pattern string) *StaticEntry {
	return &StaticEntry{
		pattern:  pattern,
		handlers: make(map[string]http.Handler),
	}
}

func (e *StaticEntry) Exec(method, str string) (http.Handler, []string) {
	if !strings.HasPrefix(str, e.pattern) {
		return nil, nil
	}
	if len(str) == len(e.pattern) {
		return e.GetHandler(method), nil
	}
	return e.traverse(method, str[len(e.pattern):])
}

func (e *StaticEntry) traverse(method, str string) (http.Handler, []string) {
	for _, entry := range e.entries {
		if h, params := entry.Exec(method, str); h != nil {
			return h, params
		}
	}
	return nil, nil
}

//
// Match
//
type MatchEntry struct {
	StaticEntry
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

	return &MatchEntry{*newStaticEntry(pat), name, matcher}
}

func (e *MatchEntry) Exec(method, str string) (http.Handler, []string) {
	i := e.matcher.Match(str)
	if i == -1 {
		return nil, nil
	}

	// finish parsing
	if len(str) == i {
		if h := e.GetHandler(method); h != nil {
			return h, []string{e.name, str}
		}
		return nil, nil
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
