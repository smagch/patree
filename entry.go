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

// type MatchEntry struct {
// 	name     string
// 	handlers map[string]*http.Handler
// 	matcher  Mather
// 	entries  []Entry
// }

// func (e *MatchEntry) Exec(str string) (*http.Handler, []string) {
// 	i, ok := e.matcher.Match(str)
// 	if !ok {
// 		return nil, nil
// 	}

// 	// finished!
// 	if len(str) == i {
// 		return nil, true
// 	}

// 	// c.str = suffix
// 	suffix := str[i:]

// 	// if suffix has still long way to go
// 	// children entry would

// 	// and then entry that capture
// 	for _, entry := range m.entries {
// 		if val, ok := entry.Exec(suffix); ok {
// 			// TODO order??
// 			// var ret = []string{}
// 			// ret = append(ret, str[])
// 			if val != nil {
// 				ret = append(ret, val...)
// 			}
// 			return ret, true
// 		}
// 	}

// 	return nil, false
// }

// func (e *StaticEntry) AddEntry(child *MatchEntry) {
// 	// entries = append(entries)
// }
