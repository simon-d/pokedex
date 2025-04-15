package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello, world  ",
			expected: []string{"hello,", "world"},
		},
		{
			input:    "HEllo    World again",
			expected: []string{"hello", "world", "again"},
		},
		{
			input:    "CharmanDer bulbasaur squIrtle",
			expected: []string{"charmander", "bulbasaur", "squirtle"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		if len(c.expected) != len(actual) {
			t.Errorf("Length of result slice did not match expected result. Expected: %d, Actual: %d", len(c.expected), len(actual))
		}

		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			if word != expectedWord {
				t.Errorf("Word result does not match expected.\n Expected: %s\n Actual: %s\n", expectedWord, word)
			}
		}
	}
}
