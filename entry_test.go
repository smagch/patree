package patree

import (
	"net/http"
	"reflect"
	"testing"
)

func foobar(w http.ResponseWriter, r *http.Request) error {
	return nil
}

var foobarHandler = HandlerFunc(foobar)
var fooValue = reflect.ValueOf(foobarHandler)

func TestStaticEntry(t *testing.T) {
	e := newStaticEntry("/foobar")
	e.SetMethodHandler("GET", foobarHandler)

	h, _ := e.exec("GET", "/foobar")
	if h == nil {
		t.Fatal("static match is broken")
	}

	h, _ = e.exec("POST", "/foobar")
	if h != nil {
		t.Fatal("cought wrong method")
	}

	h, _ = e.exec("GET", "/foobar/2000")
	if h != nil {
		t.Fatal("cought wrong path")
	}

	child := newStaticEntry("/2000")
	child.SetMethodHandler("GET", foobarHandler)
	e.AddEntry(child)
	h, _ = e.exec("GET", "/foobar/2000")
	if h == nil {
		t.Fatal("nested entry match failed")
	}

	parent := newStaticEntry("/api")
	parent.AddEntry(e)
	h, _ = parent.exec("GET", "/api/foobar/2000")
	if h == nil {
		t.Fatal("should catch nested entry")
	}
	h, _ = parent.exec("GET", "/api/foobar")
	if h == nil {
		t.Fatal("should catch nested entry")
	}
	h, _ = parent.exec("GET", "/api")
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
		h, params := e.exec("GET", s)
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

func TestMergePattern(t *testing.T) {
	mustHave := func(entry *Entry, pat string) *Entry {
		child := entry.getChildEntry(pat)
		if child == nil {
			t.Fatalf("%s must have a pattern %s\n", entry.Pattern(), pat)
		}
		return child
	}

	// 1. /foo/<int:bar>/about
	e := newStaticEntry("/foo/")
	e.MergePatterns([]string{"<int:bar>", "/about"})
	bar := mustHave(e, "<int:bar>")
	mustHave(bar, "/about")

	// 2. /foo/<int:bar>/edit/<int:edit_id>
	e.MergePatterns([]string{"<int:bar>", "/edit/", "<int:edit_id>"})
	edit := mustHave(bar, "/edit/")
	mustHave(edit, "<int:edit_id>")
	if e.Len() != 1 {
		t.Fatal("merge doesn't work properly")
	}

	// 3. /foo/<int:bar>1234567890/edit
	e.MergePatterns([]string{"<int:bar>", "1234567890", "/edit"})
	intSuffixEntry := mustHave(e, "<int:bar>1234567890")
	mustHave(intSuffixEntry, "/edit")

	intSuffixEntry.MergePatterns([]string{"<hex:hex>", "f3ab34"})
	mustHave(intSuffixEntry, "<hex:hex>f3ab34")
}

func TestOrder(t *testing.T) {
	entry := newStaticEntry("/posts/")
	testCases := []struct {
		pattern string
		index   int
	}{
		{"<int:post_id>", 5},
		{"<int:post_id>12345", 4},
		{"2014-03", 1},
		{"2014", 2},
		{"2013", 3},
		{"this-is-the-slug-of-post", 0},
	}

	for _, testCase := range testCases {
		patterns, err := SplitPath(testCase.pattern)
		if err != nil {
			t.Fatal(err)
		}
		entry.MergePatterns(patterns)
	}

	for _, testCase := range testCases {
		child := entry.entries[testCase.index]
		if child.Pattern() != testCase.pattern {
			t.Errorf("Pattern %s is at %d instead of %s\n", child.Pattern(),
				testCase.index, testCase.pattern)
		}
	}
}
