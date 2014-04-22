package patree

import (
	"net/http"
	"testing"
)

type routeTestCase struct {
	pattern string
	urlStr  string
	params  params
}

func (rtc routeTestCase) getHandler(t *testing.T) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		for k, v := range rtc.params {
			if p := Param(r, k); p != v {
				t.Fatalf("pattern %s should have param %s with url \"%s\" "+
					"when key is \"%s\". But got %s instead.", rtc.pattern,
					v, rtc.urlStr, k, p)
			}
		}
	}
	return http.HandlerFunc(h)
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
