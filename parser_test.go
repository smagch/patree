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
