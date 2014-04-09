package patree

import (
	"testing"
)

type matcherTest struct {
	m     Matcher
	cases []matcherTestCase
}

type matcherTestCase struct {
	urlStr   string
	offset   int
	matchStr string
}

func (m *matcherTest) test(t *testing.T) {
	for _, testCase := range m.cases {
		offset, matchStr := m.m.Match(testCase.urlStr)
		if offset != testCase.offset {
			t.Fatalf("Expected offset is %d with urlStr %s. Got %d instead",
				testCase.offset, testCase.urlStr, offset)
		}
		if matchStr != testCase.matchStr {
			t.Fatalf("Expected matchStr is %s with urlStr %s. Got %s instead",
				testCase.matchStr, testCase.urlStr, matchStr)
		}
	}
}

func TestIntMatcher(t *testing.T) {
	m := matcherTest{IntMatcher, []matcherTestCase{
		{"2000", 4, "2000"},
		{"1234foo", 4, "1234"},
		{"foobar", -1, ""},
		{"o091", -1, ""},
		{"091", 3, "091"},
		{"0f", 1, "0"},
		{"", -1, ""},
	}}
	m.test(t)
}

func TestHexMatcher(t *testing.T) {
	m := matcherTest{HexMatcher, []matcherTestCase{
		{"0fab3dsa0", 6, "0fab3d"},
		{"z", -1, ""},
		{"~f", -1, ""},
		{"1234", 4, "1234"},
		{"fFaa00", 6, "fFaa00"},
		{"fg", 1, "f"},
		{"Abcde0", 6, "Abcde0"},
	}}
	m.test(t)
}

func TestDefaultMatcher(t *testing.T) {
	m := matcherTest{DefaultMatcher, []matcherTestCase{
		{"foobar2000", 10, "foobar2000"},
		{"foobar/2000", 6, "foobar"},
		{"/foobar", -1, ""},
		{"日本語", len("日本語"), "日本語"},
		{"languages/にほん", 9, "languages"},
		{"日本/語", len("日本"), "日本"},
	}}
	m.test(t)
}

func TestSuffixMatcherWithIntMatcher(t *testing.T) {
	suffixMatcher := &SuffixMatcher{"-page", IntMatcher}
	m := matcherTest{suffixMatcher, []matcherTestCase{
		{"2000-page", len("2000-page"), "2000"},
		{"2000-page/edit", len("2000-page"), "2000"},
		{"1-page", len("1-page"), "1"},
		{"f-page", -1, ""},
		{"-page", -1, ""},
		{"page", -1, ""},
		{"-p", -1, ""},
		{"", -1, ""},
	}}
	m.test(t)

	// test with multibyte suffix
	suffixMatcher = &SuffixMatcher{"にほんご", IntMatcher}
	m = matcherTest{suffixMatcher, []matcherTestCase{
		{"1にほんご", len("1にほんご"), "1"},
		{"にほんご", -1, ""},
		{"あにほんご", -1, ""},
		{"234456677394023984391840132", -1, ""},
		{"432424980432897にほん", -1, ""},
		{"432424980432897にほんご", len("432424980432897にほんご"), "432424980432897"},
		{"に", -1, ""},
		{"", -1, ""},
	}}
	m.test(t)

	// test with int suffix
	suffixMatcher = &SuffixMatcher{"456789123", IntMatcher}
	m = matcherTest{suffixMatcher, []matcherTestCase{
		{"100456789123", len("100456789123"), "100"},
		{"1456789123", len("1456789123"), "1"},
		{"456789123", -1, ""},
		{"", -1, ""},
	}}
	m.test(t)
}

func TestSuffixMatcherWithDefaultMatcher(t *testing.T) {
	suffixMatcher := &SuffixMatcher{"-page", DefaultMatcher}
	m := matcherTest{suffixMatcher, []matcherTestCase{
		{"golang-page", len("golang-page"), "golang"},
		{"golang-page-staticentry", len("golang-page"), "golang"},
		{"日本語-page", len("日本語-page"), "日本語"},
		{"日本語-page/edit", len("日本語-page"), "日本語"},
		{"あ-page/edit", len("あ-page"), "あ"},
		{"-page", -1, ""},
		{"-", -1, ""},
		{"", -1, ""},
	}}
	m.test(t)

	// test with multibyte suffix
	suffixMatcher = &SuffixMatcher{"日本語", DefaultMatcher}
	m = matcherTest{suffixMatcher, []matcherTestCase{
		{"言語日本語", len("言語日本語"), "言語"},
		{"言語/日本語", -1, ""},
		{"日本語", -1, ""},
		{"", -1, ""},
	}}
}

func TestUUIDMatcher(t *testing.T) {
	m := matcherTest{UUIDMatcher, []matcherTestCase{{
		"BE567C9C-6392-4F6D-B5AE-E35893F956BB/about", 36,
		"BE567C9C-6392-4F6D-B5AE-E35893F956BB"}, {
		"e3a596d5-b7d6-4467-bf3f-d42cdac5f1be0f", 36,
		"e3a596d5-b7d6-4467-bf3f-d42cdac5f1be"}, {
		"76160D6E-924C-4A48-9739-AC12E76F176B", 36,
		"76160D6E-924C-4A48-9739-AC12E76F176B"},
		// fails
		{"", -1, ""},
		{"01754B72-", -1, ""},
		{"01754B72-ADD4-4ED9-9044-F2253CCBB85I", -1, ""},
	}}
	m.test(t)
}
