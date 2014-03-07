package treemux

import (
	"testing"
)

type matcherTest struct {
	m     Matcher
	cases map[string]int
}

func (m *matcherTest) test(t *testing.T) {
	for k, v := range m.cases {
		i := m.m.Match(k)
		if i != v {
			t.Fatalf("Got %d. Expected index is %d with argument %s", i, v, k)
		}
	}
}

func TestIntMatcher(t *testing.T) {
	m := matcherTest{IntMatcher, map[string]int{
		"2000":    4,
		"1234foo": 4,
		"foobar":  -1,
		"o091":    -1,
		"091":     3,
		"0f":      1,
		"":        -1,
	}}
	m.test(t)
}

func TestHexMatcher(t *testing.T) {
	m := matcherTest{HexMatcher, map[string]int{
		"0fab3dsa0": 6,
		"z":         -1,
		"~f":        -1,
		"1234":      4,
		"ffaa00":    6,
		"fg":        1,
	}}
	m.test(t)
}

func TestDefaultMatcher(t *testing.T) {
	m := matcherTest{DefaultMatcher, map[string]int{
		"foobar2000":    len("foobar2000"),
		"foobar/2000":   len("foobar"),
		"/foobar":       -1,
		"日本語":           len("日本語"),
		"languages/にほん": len("languages"),
		"日本/語":          len("日本"),
	}}
	m.test(t)
}

func TestSuffixMatcherWithIntMatcher(t *testing.T) {
	suffixMatcher := &SuffixMatcher{"-page", IntMatcher}
	m := matcherTest{suffixMatcher, map[string]int{
		"2000-page":      len("2000-page"),
		"2000-page/edit": len("2000-page"),
		"1-page":         len("1-page"),
		"f-page":         -1,
		"-page":          -1,
		"page":           -1,
		"-p":             -1,
		"":               -1,
	}}
	m.test(t)

	// test with multibyte suffix
	suffixMatcher = &SuffixMatcher{"にほんご", IntMatcher}
	m = matcherTest{suffixMatcher, map[string]int{
		"1にほんご":                       len("1にほんご"),
		"にほんご":                        -1,
		"あにほんご":                       -1,
		"234456677394023984391840132": -1,
		"432424980432897にほん":          -1,
		"432424980432897にほんご":         len("432424980432897にほんご"),
		"に": -1,
		"":  -1,
	}}
	m.test(t)

	// test with int suffix
	suffixMatcher = &SuffixMatcher{"456789123", IntMatcher}
	m = matcherTest{suffixMatcher, map[string]int{
		"100456789123": len("100456789123"),
		"1456789123":   len("1456789123"),
		"456789123":    -1,
		"":             -1,
	}}
	m.test(t)
}

func TestSuffixMatcherWithDefaultMatcher(t *testing.T) {
	suffixMatcher := &SuffixMatcher{"-page", DefaultMatcher}
	m := matcherTest{suffixMatcher, map[string]int{
		"golang-page":             len("golang-page"),
		"golang-page-staticentry": len("golang-page"),
		"日本語-page":                len("日本語-page"),
		"日本語-page/edit":           len("日本語-page"),
		"あ-page/edit":             len("あ-page"),
		"-page":                   -1,
		"-":                       -1,
		"":                        -1,
	}}
	m.test(t)

	// test with multibyte suffix
	suffixMatcher = &SuffixMatcher{"日本語", DefaultMatcher}
	m = matcherTest{suffixMatcher, map[string]int{
		"言語日本語":  len("言語日本語"),
		"言語/日本語": -1,
		"日本語":    -1,
		"":       -1,
	}}
}
