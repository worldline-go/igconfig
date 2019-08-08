package igconfig

import (
	"testing"
)

func TestIsTrue(t *testing.T) {
	testcases := []struct {
		c string
		w bool
	}{
		{"TRUE", true},
		{"true", true},
		{"T", true},
		{"t", true},
		{".t.", true},
		{"YeS", true},
		{"1", true},
		{"FALSE", false},
		{"treu", false},
		{"tru", false},
		{"0", false},
	}

	for _, testcase := range testcases {
		g := isTrue(testcase.c)
		if g != testcase.w {
			t.Errorf("TestIsTrue failed; got=%t; want=%t", g, testcase.w)
		}
	}
}
