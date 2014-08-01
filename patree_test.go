package patree

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type params map[string]string

type middlewareTestCase struct {
	header map[string]string
	body   string
	err    error
}

func (tc middlewareTestCase) execTests(t *testing.T) {
	count := 0
	mux := &Route{}

	Use := func(f HandlerFunc) {
		mux.Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
			count--
			f(w, r, c)
		})
		count++
	}

	Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
		c.Next(w, r)
		if c.Err != tc.err {
			t.Fatalf("Inconsistent error: %s: %s", tc.err.Error(), c.Err.Error())
		}
		if c.NotFound() {
			t.Fatalf("Should not be NotFound: %s", r.URL)
		}
	})

	// set middlewares that set a header
	for k, v := range tc.header {
		key, val := k, v
		Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
			w.Header().Set(key, val)
			c.Next(w, r)
		})
	}

	// set a middleware that write to http body
	for _, r := range tc.body {
		s := string(r)
		Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
			io.WriteString(w, s)
			c.Next(w, r)
		})
	}

	Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
		if count != 0 {
			t.Fatalf("It should have 0 counts rather than %d", count)
		}
		for k, v := range tc.header {
			if w.Header().Get(k) != v {
				t.Fatalf("Header %s should be set on %s", v, k)
			}
		}

		if v, ok := w.(*httptest.ResponseRecorder); ok {
			if body := v.Body.String(); body != tc.body {
				t.Fatal("Inconsistent body response: %s", body)
			}
		} else {
			t.Fatal("ResponseRecorder should be passed")
		}

		if tc.err != nil {
			c.Err = tc.err
		}
	})

	mux.Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
		t.Fatal("It shouldn't be called because the above middleare dosn't call",
			" c.Next(w, r)")
	})

	r, err := http.NewRequest("GET", "/middleware-test", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
}

func TestMiddleware(t *testing.T) {
	cases := []middlewareTestCase{
		{
			map[string]string{},
			"H",
			nil,
		}, {
			map[string]string{
				"X-Test":       "Test-Header",
				"Content-Type": "text/plain",
			},
			"Hello World",
			nil,
		}, {
			map[string]string{
				"Authorization": "Foobar",
				"Content-Type":  "application/json",
				"X-Test-Header": "Testing",
			},
			`{"foo": "bar"}`,
			errors.New("Middleware Error"),
		},
	}

	for _, c := range cases {
		c.execTests(t)
	}
}

type routeTestCase struct {
	pattern string
	urlStr  string
	params  params
}

func (rtc routeTestCase) getHandler(t *testing.T) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) {
		for k, v := range rtc.params {
			if p := c.Params[k]; p != v {
				t.Fatalf("pattern %s should have param %s with url \"%s\" "+
					"when key is \"%s\". But got %s instead.", rtc.pattern,
					v, rtc.urlStr, k, p)
			}
		}
	}
}

func execTests(mux *Route, cases []routeTestCase, t *testing.T) {
	mux.Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
		c.Next(w, r)
		if c.NotFound() {
			t.Fatalf("Should not be NotFound: %s", r.URL)
		}
	})

	for _, tc := range cases {
		mux.Handle(tc.pattern, tc.getHandler(t))
	}

	for _, tc := range cases {
		r, err := http.NewRequest("GET", tc.urlStr, nil)
		if err != nil {
			t.Fatal(err)
		}
		mux.ServeHTTP(nil, r)
	}
}

func TestMuxer(t *testing.T) {
	m := &Route{}

	cases := []routeTestCase{
		{"/foo/<int:bar>", "/foo/2000", params{"bar": "2000"}},
		{"/foo", "/foo", nil},
		{"/foo/<int:bar>/comments/<int:comment_id>", "/foo/1/comments/12345",
			params{"bar": "1", "comment_id": "12345"}},
		{"/foo<int:bar>", "/foo100000", params{"bar": "100000"}},
		{"/<int:bar>/token/<hex:token>", "/132/token/2afe2792458d76a7d9ff",
			params{"bar": "132", "token": "2afe2792458d76a7d9ff"}},
		{"/foo/bar<hex:token>", "/foo/bar2000f",
			params{"token": "2000f"}},
		{"/uuids/<uuid:id>", "/uuids/F2B55C6E-1B8C-4CAB-A58D-9B8DA8C31F20",
			params{"id": "F2B55C6E-1B8C-4CAB-A58D-9B8DA8C31F20"}},
		{"/uuids/<uuid:id>abcdef/<uuid:second_id>",
			"/uuids/513c96ab-b1e6-4e77-ab91-cf7dbe13a460abcdef/7B617843-065E-4F74-864C-B3B577F11D7E",
			params{"id": "513c96ab-b1e6-4e77-ab91-cf7dbe13a460",
				"second_id": "7B617843-065E-4F74-864C-B3B577F11D7E"}},
		{"/during/<date:date_start>/<date:date_end>",
			"/during/2014-01-01/2014-12-31",
			params{"date_start": "2014-01-01", "date_end": "2014-12-31"}},
		{"/date-<date:date>", "/date-2050-10-09", params{"date": "2050-10-09"}},
		{"<date:d>", "0010-05-30", params{"d": "0010-05-30"}},
	}

	execTests(m, cases, t)
}

func TestMethodMap(t *testing.T) {
	mux := &Route{}
	mux.Use(func(w http.ResponseWriter, r *http.Request, c *Context) {
		c.Next(w, r)
		if c.NotFound() {
			t.Fatalf("Should not be NotFound: %s", r.URL)
		}
		if c.Err != nil {
			t.Fatalf("Got error on %s.\n%s\n", r.URL, c.Err.Error())
		}
	})

	count := 0

	getHandlerFunc := func(pat, method string) HandlerFunc {
		count++
		return func(w http.ResponseWriter, r *http.Request, c *Context) {
			if r.URL.Path != pat {
				t.Fatalf("Cought wrong URL path '%s' for route '%s'",
					r.URL.Path, pat)
			}
			count--
		}
	}

	patterns := []string{"/handler", "/handler2", "/handler/handler2",
		"/api/1/users", "/api/1/posts", "/api/2/users", "/api/2/posts"}

	for _, pat := range patterns {
		mux.Get(pat, getHandlerFunc(pat, "GET"))
		mux.Post(pat, getHandlerFunc(pat, "POST"))
		mux.Put(pat, getHandlerFunc(pat, "PUT"))
		mux.Patch(pat, getHandlerFunc(pat, "PATCH"))
		mux.Delete(pat, getHandlerFunc(pat, "DELETE"))
		mux.Options(pat, getHandlerFunc(pat, "OPTIONS"))
	}

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	for _, pat := range patterns {
		for _, method := range methods {
			r, err := http.NewRequest(method, pat, nil)
			if err != nil {
				t.Fatal(err)
			}
			mux.ServeHTTP(nil, r)
		}
	}

	if count != 0 {
		t.Fatal("Missed executing a handler. Count should be 0 instead of", count)
	}
}
