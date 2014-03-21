package patree

import (
	"net/http"
)

type params map[string]string

var contexts = make(map[*http.Request]params)

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

type PatternTreeServeMux struct {
	StaticEntry
	notfound http.Handler
}

func New() *PatternTreeServeMux {
	return &PatternTreeServeMux{
		*newStaticEntry(""),
		http.NotFoundHandler(),
	}
}

func (m *PatternTreeServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, paramArray := m.traverse(r.Method, r.URL.Path)
	if h != nil {
		p := createParams(contexts[r], paramArray)
		contexts[r] = p
		defer delete(contexts, r)
		h.ServeHTTP(w, r)
	} else {
		m.notfound.ServeHTTP(w, r)
	}
}

func (m *PatternTreeServeMux) Handle(pat string, h http.Handler) {
	patterns := SplitPath(pat)
	leafEntry := m.MergePatterns(patterns)
	if err := leafEntry.SetHandler(h); err != nil {
		panic(err)
	}
}

func (m *PatternTreeServeMux) HandleMethod(method, pat string, h http.Handler) {
	patterns := SplitPath(pat)
	leafEntry := m.MergePatterns(patterns)
	if err := leafEntry.SetMethodHandler(method, h); err != nil {
		panic(err)
	}
}

func (m *PatternTreeServeMux) Get(pat string, h http.Handler) {
	m.HandleMethod("GET", pat, h)
	m.HandleMethod("HEAD", pat, h)
}

func (m *PatternTreeServeMux) Post(pat string, h http.Handler) {
	m.HandleMethod("POST", pat, h)
}

func (m *PatternTreeServeMux) Put(pat string, h http.Handler) {
	m.HandleMethod("PUT", pat, h)
}

func (m *PatternTreeServeMux) Patch(pat string, h http.Handler) {
	m.HandleMethod("PATCH", pat, h)
}

func (m *PatternTreeServeMux) Delete(pat string, h http.Handler) {
	m.HandleMethod("DELETE", pat, h)
}

func (m *PatternTreeServeMux) Options(pat string, h http.Handler) {
	m.HandleMethod("OPTIONS", pat, h)
}

func (m *PatternTreeServeMux) NotFound(f http.HandlerFunc) {
	m.NotFoundHandler(http.HandlerFunc(f))
}

func (m *PatternTreeServeMux) NotFoundHandler(h http.Handler) {
	m.notfound = h
}
