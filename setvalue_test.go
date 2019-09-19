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

	for i, testcase := range testcases {
		g := isTrue(testcase.c)
		if g != testcase.w {
			t.Errorf("TestIsTrue test #%d failed; got=%t; want=%t", i, g, testcase.w)
		}
	}
}

func TestSetValueWarnings(t *testing.T) {
	var c testConfig

	data, err := newLocalData(&c)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", "TestSetValueWarnings", err.Error())
	}

	tests := []struct {
		Field string
		SetTo string
	}{
		{
			Field: "Age",
			SetTo: "haha",
		},
		{
			Field: "Port",
			SetTo: "haha",
		},
		{
			Field: "Salary",
			SetTo: "haha",
		},
		{
			Field: "Unsused",
			SetTo: "1.0",
		},
	}

	for i, test := range tests {
		data.setValue(test.Field, test.SetTo)
		if data.messages == nil {
			t.Errorf("TestSetValueWarnings test #%d failed for field '%s'", i, test.Field)
		}
	}
}

func TestSetStruct(t *testing.T) {
	testStruct := struct {
		Field struct {
			Test bool
		} `cmd:"struct"`
	}{}
	data, err := newLocalData(&testStruct)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", "TestSetStruct", err.Error())
	}
	data.setValue("Field", `{"test":false}`)
	if data.messages == nil {
		t.Fail()
	}
}
