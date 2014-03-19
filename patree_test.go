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
