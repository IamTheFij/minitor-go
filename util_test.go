package main

import "testing"

func TestUtilEqualSliceString(t *testing.T) {
	cases := []struct {
		a, b     []string
		expected bool
	}{
		// Equal cases
		{nil, nil, true},
		{nil, []string{}, true},
		{[]string{}, nil, true},
		{[]string{"a"}, []string{"a"}, true},
		// Inequal cases
		{nil, []string{"b"}, false},
		{[]string{"a"}, nil, false},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a"}, []string{"a", "b"}, false},
		{[]string{"a", "b"}, []string{"b"}, false},
	}

	for _, c := range cases {
		actual := EqualSliceString(c.a, c.b)
		if actual != c.expected {
			t.Errorf(
				"EqualSliceString(%v, %v), expected=%v actual=%v",
				c.a, c.b, c.expected, actual,
			)
		}
	}
}
