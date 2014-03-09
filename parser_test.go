package treemux

import (
	"reflect"
	"testing"
)

func TestSplitPath(t *testing.T) {
	cases := map[string][]string{
		"/foo/bar/2000":             {"/foo/", "bar/", "2000"},
		"/foo/<int:bar>":            {"/foo/", "<int:bar>"},
		"/foo/<int:bar>/":           {"/foo/", "<int:bar>", "/"},
		"/foo/<int:bar>/edit":       {"/foo/", "<int:bar>", "/edit"},
		"/foo/<int:bar>-page":       {"/foo/", "<int:bar>", "-page"},
		"/foo/<int:bar>-page/":      {"/foo/", "<int:bar>", "-page/"},
		"/foo/<int:bar>-page/about": {"/foo/", "<int:bar>", "-page/", "about"},
	}
	for p, expected := range cases {
		ret := SplitPath(p)
		if !reflect.DeepEqual(ret, expected) {
			t.Fatalf("Got %v instead of expected %v with input %s", ret, expected, p)
		}
	}
}

func TestNewEntry(t *testing.T) {
	cases := map[string]Entry{
		"/foo/":          newStatic("/foo/"),
		"foo/":           newStatic("foo/"),
		"<int:foo>":      newMatchEntry("foo", IntMatcher),
		"<int:>":         newMatchEntry("", IntMatcher),
		"<foo>":          newMatchEntry("foo", DefaultMatcher),
		"<:foo>":         newMatchEntry("foo", DefaultMatcher),
		"<hex:bar>":      newMatchEntry("bar", HexMatcher),
		"<default:hoge>": newMatchEntry("hoge", DefaultMatcher),
	}
	for s, e := range cases {
		entry, err := NewEntry(s)
		if err != nil {
			t.Fatal(err.Error())
		}
		if !e.Equals(entry) {
			t.Fatalf("Got %v instead of Expected %v", e, entry)
		}
	}
}
