package patree

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
		"/<int:bar>/about":          {"/", "<int:bar>", "/about"},
	}
	for p, expected := range cases {
		ret, err := SplitPath(p)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(ret, expected) {
			t.Fatalf("Got %v instead of expected %v with input %s", ret, expected, p)
		}
	}

	errorCases := []string{
		"/foo/<bar:", "/posts/<int", "/posts/<int:post_id", "<int",
	}

	for _, p := range errorCases {
		_, err := SplitPath(p)
		if err == nil {
			t.Fatalf("it should have error with pattern %s\n", p)
		}
		if err != NoClosingBracket {
			t.Fatal(err)
		}
	}
}
