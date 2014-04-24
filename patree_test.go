package patree

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type routeTestCase struct {
	pattern string
	urlStr  string
	params  params
}

func (rtc routeTestCase) getHandler(t *testing.T) Handler {
	h := func(w http.ResponseWriter, r *http.Request) error {
		for k, v := range rtc.params {
			if p := Param(r, k); p != v {
				t.Fatalf("pattern %s should have param %s with url \"%s\" "+
					"when key is \"%s\". But got %s instead.", rtc.pattern,
					v, rtc.urlStr, k, p)
			}
		}
		return nil
	}
	return HandlerFunc(h)
}

func execTests(m *PatternTreeServeMux, cases []routeTestCase, t *testing.T) {
	m.NotFoundFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("Should not be NotFound: %s", r.URL)
	})
	for _, c := range cases {
		m.Handle(c.pattern, c.getHandler(t))
		r, err := http.NewRequest("GET", c.urlStr, nil)
		if err != nil {
			t.Fatal(err)
		}
		m.ServeHTTP(nil, r)
	}
}

type middlewareTestCase struct {
	header map[string]string
	body   string
	err    error
}

func (c middlewareTestCase) execTests(t *testing.T) {
	m := New()
	m.NotFoundFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("Should not be NotFound: %s", r.URL)
	})

	var total, count int

	// set middlewares that set a header
	for k, v := range c.header {
		f := func(k, v string) HandlerFunc {
			f := func(w http.ResponseWriter, r *http.Request) error {
				w.Header().Set(k, v)
				return nil
			}
			count++
			return f
		}(k, v)
		m.UseFunc(f)
		total++
	}

	// set a middleware that write to http body
	handlers := []HandlerFunc{}
	for _, r := range c.body {
		s := string(r)
		f := func(w http.ResponseWriter, r *http.Request) error {
			io.WriteString(w, s)
			count++
			return nil
		}
		handlers = append(handlers, f)
		total++
	}

	validate := func(w http.ResponseWriter, r *http.Request) error {
		if count != total {
			t.Fatalf("It should have %d counts rather than %d", total, count)
		}
		for k, v := range c.header {
			if w.Header().Get(k) != v {
				t.Fatalf("Header %s should be set on %s", v, k)
			}
		}

		if v, ok := w.(*httptest.ResponseRecorder); ok {
			if body := v.Body.String(); body != c.body {
				t.Fatal("Inconsistent body response: %s", body)
			}
		} else {
			t.Fatal("ResponseRecorder should be passed")
		}

		return nil
	}

	if c.err != nil {
		fError := func(w http.ResponseWriter, r *http.Request) error {
			count++
			return c.err
		}
		total++
		f := func(w http.ResponseWriter, r *http.Request) error {
			t.Fatal("Error Handler should be called instead")
			return nil
		}
		handlers = append(handlers, fError, f)
		m.ErrorFunc(func(w http.ResponseWriter, r *http.Request, err error) {
			if err != c.err {
				t.Fatalf("Should catch an error %v", err)
			}
			count++
			validate(w, r)
		})
		total++
	} else {
		handlers = append(handlers, validate)
	}

	m.HandleFunc("/middleware-test", handlers...)

	r, err := http.NewRequest("GET", "/middleware-test", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)
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

func TestMuxer(t *testing.T) {
	m := New()

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
	}

	execTests(m, cases, t)
}

func TestPrefixMuxer(t *testing.T) {
	m := NewWithPrefix("/api/1")

	cases := []routeTestCase{
		{"/posts/<int:id>", "/api/1/posts/1003", params{"id": "1003"}},
		{"/posts", "/api/1/posts", nil},
		{"/<int:id>/comments/<hex:comment_id>", "/api/1/10/comments/0f3d51",
			params{"id": "10", "comment_id": "0f3d51"}},
	}

	execTests(m, cases, t)
}

func TestMethodMap(t *testing.T) {
	m := New()
	m.NotFoundFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("Should not be NotFound: %s", r.URL)
	})
	m.ErrorFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		t.Fatalf("Should not have an error: %s", r.URL, err.Error())
	})

	var count, total int

	getHandlerFunc := func(pat, method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			if r.URL.Path != pat {
				t.Fatalf("Cought wrong URL path '%s' for route '%s'",
					r.URL.Path, pat)
			}
			count++
			return nil
		}
	}

	getHandler := func(pat, method string) Handler {
		f := getHandlerFunc(pat, method)
		return f
	}

	pattern := "/handler"
	m.Get(pattern, getHandler(pattern, "GET"))
	m.Post(pattern, getHandler(pattern, "POST"))
	m.Put(pattern, getHandler(pattern, "PUT"))
	m.Patch(pattern, getHandler(pattern, "PATCH"))
	m.Delete(pattern, getHandler(pattern, "DELETE"))
	m.Options(pattern, getHandler(pattern, "OPTIONS"))

	pattern = "/handler-func"
	m.GetFunc(pattern, getHandlerFunc(pattern, "GET"))
	m.PostFunc(pattern, getHandlerFunc(pattern, "POST"))
	m.PutFunc(pattern, getHandlerFunc(pattern, "PUT"))
	m.PatchFunc(pattern, getHandlerFunc(pattern, "PATCH"))
	m.DeleteFunc(pattern, getHandlerFunc(pattern, "DELETE"))
	m.OptionsFunc(pattern, getHandlerFunc(pattern, "OPTIONS"))

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	patterns := []string{"/handler", "/handler-func"}
	for _, method := range methods {
		for _, pat := range patterns {
			r, err := http.NewRequest(method, pat, nil)
			if err != nil {
				t.Fatal(err)
			}
			total++
			m.ServeHTTP(nil, r)
		}
	}

	if total != count {
		t.Fatal("Missed executing a handler")
	}
}
