package patree

import (
	"net/http"
)

var context = make(map[*http.Request][]string)

func GetParams(r *http.Request) map[string]string {
	params := context[r]
	paramMap := make(map[string]string)
	for i := len(params) - 1; i >= 1; i -= 2 {
		paramMap[params[i-1]] = params[i]
	}
	return paramMap
}

func GetParamValues(r *http.Request) (values []string) {
	params := context[r]
	for i := len(params) - 1; i >= 1; i -= 2 {
		values = append(values, params[i])
	}
	return
}

func New() *PatternTreeServeMux {
	return &PatternTreeServeMux{*newStaticEntry("")}
}

type PatternTreeServeMux struct {
	StaticEntry
}

func (m *PatternTreeServeMux) Handle(pat string, h http.Handler) {
	patterns := SplitPath(pat)
	leafEntry := m.MergePatterns(patterns)
	leafEntry.SetHandler(h)
}

func (m *PatternTreeServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, params := m.traverse(r.Method, r.URL.Path)
	if h != nil {
		// in case nested use case
		context[r] = append(params, context[r]...)
		defer delete(context, r)
		h.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}
