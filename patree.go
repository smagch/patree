// package patree implements a simple http request multiplexer.
package patree

import (
	"net/http"
)

// params is a map of matched request parameters. For example, a request with
// url "/posts/2345" will have "post_id" key and "2345" value with pattern
// "/posts/<int:post_id>".
type params map[string]string

// contexts stores params for http requests.
var contexts = make(map[*http.Request]params)

// Param returns a mathced request parameter with the given key.
func Param(r *http.Request, key string) string {
	return contexts[r][key]
}

func createParams(p params, paramArray []string) params {
	if p == nil {
		p = make(params)
	}
	for i := len(paramArray) - 1; i >= 1; i -= 2 {
		p[paramArray[i-1]] = paramArray[i]
	}
	return p
}

// PatternTreeServeMux is an HTTP request multiplexer that does pattern matching.
type PatternTreeServeMux struct {
	rootEntry *Entry
	notfound  http.Handler
}

// New creates a new muxer
func New() *PatternTreeServeMux {
	entry := newStaticEntry("")
	entry.exec = entry.traverse
	return &PatternTreeServeMux{
		rootEntry: entry,
		notfound:  http.NotFoundHandler(),
	}
}

// ServeHTTP execute matched handler or execute notfound handler.
func (m *PatternTreeServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, paramArray := m.rootEntry.exec(r.Method, r.URL.Path)
	if h != nil {
		p := createParams(contexts[r], paramArray)
		contexts[r] = p
		defer delete(contexts, r)
		h.ServeHTTP(w, r)
	} else {
		m.notfound.ServeHTTP(w, r)
	}
}

// Handle a http.Handler with the given url pattern. panic with duplicate
// handler registration for a pattern.
func (m *PatternTreeServeMux) Handle(pat string, h http.Handler) {
	patterns, err := SplitPath(pat)
	if err != nil {
		panic(err)
	}
	leafEntry := m.rootEntry.MergePatterns(patterns)
	err = leafEntry.SetHandler(h)
	if err != nil {
		panic(err)
	}
}

// Handle the handler with the given http method and pattern. Panic with
// duplicate handler registration.
func (m *PatternTreeServeMux) HandleMethod(method, pat string, h http.Handler) {
	patterns, err := SplitPath(pat)
	if err != nil {
		panic(err)
	}
	leafEntry := m.rootEntry.MergePatterns(patterns)
	err = leafEntry.SetMethodHandler(method, h)
	if err != nil {
		panic(err)
	}
}

// Get registers the handler with the given pattern for "GET" and "HEAD" method.
func (m *PatternTreeServeMux) Get(pat string, h http.Handler) {
	m.HandleMethod("GET", pat, h)
	m.HandleMethod("HEAD", pat, h)
}

// Post registers the handler with the given pattern for "POST" method.
func (m *PatternTreeServeMux) Post(pat string, h http.Handler) {
	m.HandleMethod("POST", pat, h)
}

// Put registers the handler with the given pattern for "PUT" method.
func (m *PatternTreeServeMux) Put(pat string, h http.Handler) {
	m.HandleMethod("PUT", pat, h)
}

// Patch registers the handler with the given pattern for "PATCH" method.
func (m *PatternTreeServeMux) Patch(pat string, h http.Handler) {
	m.HandleMethod("PATCH", pat, h)
}

// Delete registers the handler with the give pattern for "DELETE" method.
func (m *PatternTreeServeMux) Delete(pat string, h http.Handler) {
	m.HandleMethod("DELETE", pat, h)
}

// Options registers the handler with the given pattern for "OPTIONS" method.
func (m *PatternTreeServeMux) Options(pat string, h http.Handler) {
	m.HandleMethod("OPTIONS", pat, h)
}

// NotFound registers fallback HandlerFunc in case no pattern matches.
func (m *PatternTreeServeMux) NotFound(f http.HandlerFunc) {
	m.NotFoundHandler(http.HandlerFunc(f))
}

// NotFoundHandler reigsters fallback Handler in case no pattern matches.
func (m *PatternTreeServeMux) NotFoundHandler(h http.Handler) {
	m.notfound = h
}
