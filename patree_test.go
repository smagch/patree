package patree

import (
	"net/http"
	"reflect"
	"testing"
)

func TestParameters(t *testing.T) {
	m := New()
	foobar := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	m.Handle("/foo/<int:bar>", foobar)
	m.Handle("/foo/<int:bar>/comments", foobar)
	m.Handle("/foo/<int:bar>/comments/<int:comment_id>", foobar)
	m.Handle("/foo/bar", foobar)
	cases := map[string][]string{
		"/foo/2000":            {"bar", "2000"},
		"/foo/103493204":       {"bar", "103493204"},
		"/foo/1":               {"bar", "1"},
		"/foo/bar":             nil,
		"/foo/2000/comments":   {"bar", "2000"},
		"/foo/1234/comments/0": {"comment_id", "0", "bar", "1234"},
	}

	for url, params := range cases {
		h, p := m.traverse("GET", url)
		if h == nil {
			t.Fatalf("handler should exist")
		}
		if !reflect.DeepEqual(params, p) {
			t.Fatalf("params should be %v. But got %v\n", params, p)
		}
	}

	// notfounds
	for _, url := range []string{"/foo/", "/foo", "/", "/foo/hoge"} {
		if h, p := m.traverse("GET", url); h != nil || p != nil {
			t.Fatalf("URL \"%s\" should be notfound")
		}
	}
}

func TestHandlers(t *testing.T) {
	m := New()
	m.NotFound(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("Should not be NotFound: %s", r.URL)
	})

	type ctx struct {
		method  string
		pattern string
		urlStr  string
		params  map[string]string
	}

	getHandler := func(c *ctx) http.Handler {
		h := func(w http.ResponseWriter, r *http.Request) {
			for k, v := range c.params {
				if p := Param(r, k); p != v {
					t.Fatalf("pattern %s should have param %s with url \"%s\" "+
						"when key is \"%s\". But got %s instead.", c.pattern,
						v, c.urlStr, k, p)
				}
			}
		}
		return http.HandlerFunc(h)
	}

	cases := []ctx{
		{"GET", "/foo/<int:bar>", "/foo/2000", params{"bar": "2000"}},
		{"POST", "/foo", "/foo", nil},
		{"PUT", "/foo/<int:bar>/comments/<int:comment>",
			"/foo/1/comments/12345", params{"bar": "1", "comment": "12345"}},
		{"PATCH", "/foo/<int:bar>", "/foo/100000", nil},
		{"DELETE", "/<int:bar>/token/<hex:token>",
			"/132/token/2afe2792458d76a7d9ff",
			params{"bar": "132", "token": "2afe2792458d76a7d9ff"}},
		{"OPTIONS", "/foo", "/foo", nil},
	}

	for _, c := range cases {
		m.HandleMethod(c.method, c.pattern, getHandler(&c))
	}

	for _, c := range cases {
		r, err := http.NewRequest(c.method, c.urlStr, nil)
		if err != nil {
			t.Fatal(err)
		}
		m.ServeHTTP(nil, r)
	}
}
