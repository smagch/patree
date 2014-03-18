package patree

import (
	"net/http"
	"reflect"
	"testing"
)

func foobar(w http.ResponseWriter, r *http.Request) {}

var foobarHandler = http.HandlerFunc(foobar)
var fooValue = reflect.ValueOf(foobarHandler)

func TestStaticEntry(t *testing.T) {
	e := newStaticEntry("/foobar")
	e.handlers["GET"] = foobarHandler

	h, _ := e.Exec("GET", "/foobar")
	if h == nil {
		t.Fatal("static match is broken")
	}

	h, _ = e.Exec("POST", "/foobar")
	if h != nil {
		t.Fatal("cought wrong method")
	}

	h, _ = e.Exec("GET", "/foobar/2000")
	if h != nil {
		t.Fatal("cought wrong path")
	}

	child := newStaticEntry("/2000")
	child.handlers["GET"] = foobarHandler
	e.add(child)
	h, _ = e.Exec("GET", "/foobar/2000")
	if h == nil {
		t.Fatal("nested entry match failed")
	}

	parent := newStaticEntry("/api")
	parent.add(e)
	h, _ = parent.Exec("GET", "/api/foobar/2000")
	if h == nil {
		t.Fatal("should catch nested entry")
	}
	h, _ = parent.Exec("GET", "/api/foobar")
	if h == nil {
		t.Fatal("should catch nested entry")
	}
	h, _ = parent.Exec("GET", "/api")
	if h != nil {
		t.Fatal("should not catch nested entry")
	}
}

func TestIntMatchEntry(t *testing.T) {
	cases := map[string]bool{
		"2345":               true,
		"123f":               false,
		"a":                  false,
		"0245Z245":           false,
		"139093449850284011": true,
	}

	e := newMatchEntry("<int:test_id>")
	e.handlers["GET"] = foobarHandler

	for s, ok := range cases {
		h, params := e.Exec("GET", s)
		if !ok && h != nil {
			t.Fatal("\"%s\" should return nil handler", s)
		}
		if !ok && params != nil {
			t.Fatal("\"%s\" should return nil parameters", params)
		}
		if ok && reflect.ValueOf(h).Pointer() != fooValue.Pointer() {
			t.Fatal("handler should have same pointer")
		}
		if ok && params[0] != "test_id" {
			t.Fatal("the first parameter should be a name of match")
		}
		if ok && params[1] != s {
			t.Fatal("the second parameter should be a matched result")
		}
	}
}

// func TestNewEntry(t *testing.T) {
// 	cases := make(map[string]Entry)
// 	cases["/foo/":          newStatic("/foo/"),
// 		"foo/":           newStatic("foo/"),
// 		"<int:foo>":      newMatchEntry("foo", IntMatcher),
// 		"<int:>":         newMatchEntry("", IntMatcher),
// 		"<foo>":          newMatchEntry("foo", DefaultMatcher),
// 		"<:foo>":         newMatchEntry("foo", DefaultMatcher),
// 		"<hex:bar>":      newMatchEntry("bar", HexMatcher),
// 		"<default:hoge>": newMatchEntry("hoge", DefaultMatcher),
// 	}
// 	for s, e := range cases {
// 		entry, err := NewEntry(s)
// 		if err != nil {
// 			t.Fatal(err.Error())
// 		}
// 		if !e.Equals(entry) {
// 			t.Fatalf("Got %v instead of Expected %v", e, entry)
// 		}
// 	}
// }

// func TestSuffixMatchEntry(t *testing.T) {
// 	SuffixMatcher := &SuffixMatcher{"-page", DefaultMatcher}
// 	e := newSuffixMatchEntry("pager", SuffixMatcher)
// 	e.handlers["GET"] = foobarHandler

// 	cases := map[string][]string{
// 		"234565-page": []string{"pager", "234565"},
// 		"100-page":    []string{"pager", "100"},
// 		"1-page":      []string{"pager", "1"},
// 		"世界-page":     []string{"pager", "世界"},
// 		"-page":       nil,
// 		"":            nil,
// 	}

// 	for s, params := range cases {
// 		_, p := e.Exec("GET", s)
// 		if !reflect.DeepEqual(params, p) {
// 			t.Fatalf("param should be %v. But got %v\n", params, p)
// 		}
// 		if h, _ := e.Exec("POST", s); h != nil {
// 			t.Fatal("It shouldn't catch POST method")
// 		}
// 	}
// }
