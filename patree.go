// package patree implements a simple http request multiplexer.
package patree

import (
	"net/http"
)

// Handler is a http Handler with context
type Handler interface {
	ServeHTTPContext(w http.ResponseWriter, r *http.Request, c *Context)
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as
// patree handlers.
type HandlerFunc func(http.ResponseWriter, *http.Request, *Context)

// ServeHTTPContext calls f(w, r, c)
func (f HandlerFunc) ServeHTTPContext(w http.ResponseWriter, r *http.Request, c *Context) {
	f(w, r, c)
}

// Context represents a context of http request.
type Context struct {
	route  *Route
	Params map[string]string
	Err    error
	holdUp bool
}

// Next invoke next route with the given ResponseWriter and Request
func (c *Context) Next(w http.ResponseWriter, r *http.Request) {
	if c.route == nil {
		return
	}

	if next := c.route.next; next != nil {
		c.route = next
		next.ServeHTTPContext(w, r, c)
	} else {
		c.holdUp = true
	}
}

// Notfound returns a boolean if the context is consumed all routes.
func (c *Context) NotFound() bool {
	return c.holdUp
}

type patternRouter struct {
	entry *Entry
}

func newRouter() *patternRouter {
	entry := newStaticEntry("")
	entry.exec = entry.traverse
	return &patternRouter{entry}
}

func (p *patternRouter) ServeHTTPContext(w http.ResponseWriter, r *http.Request, c *Context) {
	route, paramArray := p.entry.exec(r.Method, r.URL.Path)
	if route == nil {
		c.Next(w, r)
		return
	}

	// TODO hold old maps
	c.Params = createParams(paramArray)

	current := c.route
	c.route = route
	route.ServeHTTPContext(w, r, c)
	c.route = current

	if c.holdUp {
		c.holdUp = false
		c.Next(w, r)
	}
}

func (p *patternRouter) registerPattern(pat string) *Entry {
	patterns, err := SplitPath(pat)
	if err != nil {
		panic(err)
	}
	return p.entry.MergePatterns(patterns)
}

// Route is a chainable handler
type Route struct {
	f    Handler
	next *Route
}

// ServeHTTP implement http.Handler interface
func (route *Route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &Context{route: route}
	route.ServeHTTPContext(w, r, c)
}

// ServeHTTPContext implements Handler interface
func (route *Route) ServeHTTPContext(w http.ResponseWriter, r *http.Request, c *Context) {
	route.f.ServeHTTPContext(w, r, c)
}

// Use appends a HandlerFunc to the route.
func (r *Route) Use(f HandlerFunc) {
	r.UseHandler(f)
}

// UseHandler appends a Handler to the route.
func (r *Route) UseHandler(h Handler) {
	if r.f == nil {
		r.f = h
		return
	}

	route := r.getLeaf()
	route.next = &Route{f: h}
}

func (r *Route) getLeaf() *Route {
	if r.f == nil {
		return r
	}
	route := r
	for route.next != nil {
		route = route.next
	}
	return route
}

func (r *Route) addPattern(pat string) *Entry {
	var p *patternRouter
	var isRouter bool
	route := r.getLeaf()

	if route.f != nil {
		p, isRouter = route.f.(*patternRouter)
	}

	if !isRouter {
		p = newRouter()
		defer r.UseHandler(p)
	}

	entry := p.registerPattern(pat)
	return entry
}

// HandleMethod registers handler funcs with the given pattern and method.
func (r *Route) HandleMethod(pat, method string, f ...HandlerFunc) {
	entry := r.addPattern(pat)
	batch := batchRoute(f)
	if err := entry.SetMethodHandler(method, batch); err != nil {
		panic(err)
	}
}

// Handle registers handler funcs with the given pattern.
func (r *Route) Handle(pat string, f ...HandlerFunc) {
	entry := r.addPattern(pat)
	batch := batchRoute(f)
	if err := entry.SetHandler(batch); err != nil {
		panic(err)
	}
}

// Get registers handlers with the given pattern for GET and HEAD method
func (r *Route) Get(pat string, f ...HandlerFunc) {
	r.HandleMethod(pat, "GET", f...)
	r.HandleMethod(pat, "HEAD", f...)
}

// Post registers handlers with the given pattern for POST method
func (r *Route) Post(pat string, f ...HandlerFunc) {
	r.HandleMethod(pat, "POST", f...)
}

// Put registers handlers with the given pattern for PUT method
func (r *Route) Put(pat string, f ...HandlerFunc) {
	r.HandleMethod(pat, "PUT", f...)
}

// Patch registers handlers with the given pattern for PATCH method
func (r *Route) Patch(pat string, f ...HandlerFunc) {
	r.HandleMethod(pat, "PATCH", f...)
}

// Delete registers handlers with the given pattern for DELETE method
func (r *Route) Delete(pat string, f ...HandlerFunc) {
	r.HandleMethod(pat, "DELETE", f...)
}

// Options registers handlers with the given pattern for OPTIONS method
func (r *Route) Options(pat string, f ...HandlerFunc) {
	r.HandleMethod(pat, "OPTIONS", f...)
}

// TODO refactor Entry and drop this
func createParams(paramArray []string) map[string]string {
	p := make(map[string]string)
	for i := len(paramArray) - 1; i >= 1; i -= 2 {
		p[paramArray[i-1]] = paramArray[i]
	}
	return p
}

func batchRoute(f []HandlerFunc) *Route {
	batch := &Route{}
	for _, h := range f {
		batch.Use(h)
	}
	return batch
}
