package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
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
		{"1", true},
		{"FALSE", false},
		{"treu", false},
		{"tru", false},
		{"0", false},
	}

	for i, testcase := range testcases {
		g := isTrue(testcase.c)
		if g != testcase.w {
			t.Errorf("TestIsTrue test #%d failed; got=%t; want=%t", i, g, testcase.w)
		}
	}
}

func TestSetValueWarnings(t *testing.T) {
	var c testdata.TestConfig

	tests := []struct {
		Field string
		Val   string
	}{
		{
			Field: "Age",
			Val:   "haha",
		},
		{
			Field: "Port",
			Val:   "haha",
		},
		{
			Field: "Salary",
			Val:   "haha",
		},
		{
			Field: "Unsused",
			Val:   "1.0",
		},
	}

	for i, test := range tests {
		refVal, _ := GetReflectElem(&c)
		assert.NotNil(t, SetStructFieldValue(test.Field, test.Val, refVal), fmt.Sprintf("test #%d", i))
	}
}

func TestSetStruct(t *testing.T) {
	testStruct := struct {
		Field struct {
			Test bool
		} `cmd:"struct"`
	}{}

	refVal, _ := GetReflectElem(&testStruct)
	assert.NotNil(t, SetStructFieldValue("Field", `{"test":false`, refVal))
}
