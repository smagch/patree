package treemux

import (
	"testing"
)

func TestIntMatcher(t *testing.T) {
	m := IntMatcher{}
	cases := map[string]int{
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
