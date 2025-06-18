package wildcard

import "testing"

func TestWildCard(t *testing.T) {
	testCases := []struct {
		pattern string
		str     string
		match   bool
	}{
		{"a", "a", true},
		{"a", "b", false},
		{"a*", "abc", true},
	}

	for _, tc := range testCases {
		pattern, err := Compile(tc.pattern)
		if err != nil {
			t.Errorf("Failed to compile pattern %s: %v", tc.pattern, err)
			continue
		}
		if pattern.Match(tc.str) != tc.match {
			t.Errorf("Pattern %s did not match string %s as expected", tc.pattern, tc.str)
		}
	}
}
