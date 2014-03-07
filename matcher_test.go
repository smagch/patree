package treemux

import (
	"testing"
)

func TestIntMatcher(t *testing.T) {
	m := IntMatcher
	cases := map[string]int {
		"2000": 4,
		"1234foo": 4,
		"foobar": -1,
		"o091": -1,
		"091": 3,
		"0f": 1,
		"": -1,
	}

	for k, v := range cases {
		i := m.Match(k)
		if i != v {
			t.Fatalf("Got %d. Expected index is %d with argument %s", i, v, k)
		}
	}
}

func TestHexMatcher(t *testing.T) {
	m := HexMatcher
	cases := map[string]int {
		"0fab3dsa0": 6,
		"z": -1,
		"~f": -1,
		"1234": 4,
		"ffaa00": 6,
		"fg": 1,
	}
	for k, v := range cases {
		i := m.Match(k)
		if i != v {
			t.Fatalf("Got %d. Expected index is %d with argument %s", i, v, k)
		}
	}
}

func TestDefaultMatcher(t *testing.T) {
	m := DefaultMatcher
	cases := map[string]int {
		"foobar2000": len("foobar2000"),
		"foobar/2000": len("foobar"),
		"/foobar": -1,
		"日本語": len("日本語"),
		"languages/にほん": len("languages"),
		"日本/語": len("日本"),
	}
	for k, v := range cases {
		i := m.Match(k)
		if i != v {
			t.Fatalf("Got %d. Expected index is %d with argument %s", i, v, k)
		}
	}
}

func TestSuffixMatcherWithIntMatcher(t *testing.T) {
	m := SuffixMatcher{"-page", IntMatcher}
	cases := map[string]int{
		"2000-page":      len("2000-page"),
		"2000-page/edit": len("2000-page"),
		"1-page":         len("1-page"),
		"f-page":         -1,
		"-page":          -1,
		"page":           -1,
		"-p":             -1,
	}
	for k, v := range cases {
		i := m.Match(k)
		if i != v {
			t.Fatalf("Got %d. Expected index is %d with argument %s", i, v, k)
		}
	}
}
