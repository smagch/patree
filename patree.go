// package patree implements a simple http request multiplexer.
package patree

import (
	"net/http"
)

// Handler is a http Handler that is able to return an error
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// Handler is a handler function that have an error return
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP calls f(w, r)
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// ErrorHandler is a http Handler that have an error in the third argument.
type ErrorHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, err error)
}

// ErrorHandlerFunc
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

// ServeHTTP calls f(w, r, err)
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}

// Middleware
type Middleware []Handler

// ServeHTTP implements a patree.Handler interface
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	for _, h := range m {
		if err := h.ServeHTTP(w, r); err != nil {
			return err
		}
	}
	return nil
}

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

// ErrorHandler is the default ErrorHandler
func HandlerError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// PatternTreeServeMux is an HTTP request multiplexer that does pattern matching.
type PatternTreeServeMux struct {
	rootEntry    *Entry
	middleware   Middleware
	notfound     http.Handler
	errorHandler ErrorHandler
}

// New creates a new muxer
func New() *PatternTreeServeMux {
	entry := newStaticEntry("")
	entry.exec = entry.traverse
	return &PatternTreeServeMux{
		rootEntry:    entry,
		notfound:     http.NotFoundHandler(),
		errorHandler: ErrorHandlerFunc(HandlerError),
	}
}

// NewWithPrefix create a new muxer with the given prefix pattern string.
func NewWithPrefix(pat string) *PatternTreeServeMux {
	entry := newStaticEntry(pat)
	return &PatternTreeServeMux{
		rootEntry: entry,
		notfound:  http.NotFoundHandler(),
	}
}

// ServeHTTP execute matched handler or execute notfound handler.
func (m *PatternTreeServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, paramArray := m.rootEntry.exec(r.Method, r.URL.Path)
	if h == nil {
		m.notfound.ServeHTTP(w, r)
		return
	}

	p := createParams(contexts[r], paramArray)
	contexts[r] = p
	defer delete(contexts, r)
	if err := m.middleware.ServeHTTP(w, r); err != nil {
		m.errorHandler.ServeHTTP(w, r, err)
		return
	}
	if err := h.ServeHTTP(w, r); err != nil {
		m.errorHandler.ServeHTTP(w, r, err)
	}
}

// Use appens a Handler as Middleware
func (m *PatternTreeServeMux) Use(h Handler) {
	m.middleware = append(m.middleware, h)
}

// UseFunc appends a HandlerFunc as middleware
func (m *PatternTreeServeMux) UseFunc(h HandlerFunc) {
	m.middleware = append(m.middleware, h)
}

// registerPattern merge pattern with the given handlers
func (m *PatternTreeServeMux) registerPattern(pat string, h []Handler) (Handler, *Entry) {
	patterns, err := SplitPath(pat)
	if err != nil {
		panic(err)
	}
	var handler Handler
	if len(h) > 1 {
		handler = Middleware(h)
	} else {
		handler = h[0]
	}
	leafEntry := m.rootEntry.MergePatterns(patterns)
	return handler, leafEntry
}

// HandleFunc add Handlers with the given pattern.
func (m *PatternTreeServeMux) HandleFunc(pat string, h ...HandlerFunc) {
	handlers := make([]Handler, len(h))
	for i, f := range h {
		handlers[i] = HandlerFunc(f)
	}
	m.Handle(pat, handlers...)
}

// Handle a http.Handler with the given url pattern. panic with duplicate
// handler registration for a pattern.
func (m *PatternTreeServeMux) Handle(pat string, h ...Handler) {
	handler, leafEntry := m.registerPattern(pat, h)
	if err := leafEntry.SetHandler(handler); err != nil {
		panic(err)
	}
}

// Handle the handler with the given http method and pattern. Panic with
// duplicate handler registration.
func (m *PatternTreeServeMux) HandleMethod(method, pat string, h ...Handler) {
	handler, leafEntry := m.registerPattern(pat, h)
	if err := leafEntry.SetMethodHandler(method, handler); err != nil {
		panic(err)
	}
}

// Get registers the handler with the given pattern for "GET" and "HEAD" method.
func (m *PatternTreeServeMux) Get(pat string, h ...Handler) {
	m.HandleMethod("GET", pat, h...)
	m.HandleMethod("HEAD", pat, h...)
}

// Post registers the handler with the given pattern for "POST" method.
func (m *PatternTreeServeMux) Post(pat string, h ...Handler) {
	m.HandleMethod("POST", pat, h...)
}

// Put registers the handler with the given pattern for "PUT" method.
func (m *PatternTreeServeMux) Put(pat string, h ...Handler) {
	m.HandleMethod("PUT", pat, h...)
}

// Patch registers the handler with the given pattern for "PATCH" method.
func (m *PatternTreeServeMux) Patch(pat string, h ...Handler) {
	m.HandleMethod("PATCH", pat, h...)
}

// Delete registers the handler with the give pattern for "DELETE" method.
func (m *PatternTreeServeMux) Delete(pat string, h ...Handler) {
	m.HandleMethod("DELETE", pat, h...)
}

// Options registers the handler with the given pattern for "OPTIONS" method.
func (m *PatternTreeServeMux) Options(pat string, h ...Handler) {
	m.HandleMethod("OPTIONS", pat, h...)
}

// NotFoundFunc registers fallback HandlerFunc in case no pattern matches.
func (m *PatternTreeServeMux) NotFoundFunc(f http.HandlerFunc) {
	m.NotFound(f)
}

// NotFound reigsters fallback Handler in case no pattern matches.
func (m *PatternTreeServeMux) NotFound(h http.Handler) {
	m.notfound = h
}

// ErrorFunc registers an error handler function.
func (m *PatternTreeServeMux) ErrorFunc(h ErrorHandlerFunc) {
	m.Error(h)
}

// Error registers an error handler.
func (m *PatternTreeServeMux) Error(h ErrorHandler) {
	m.errorHandler = h
}
