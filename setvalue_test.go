package igconfig

import (
	"reflect"
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

func TestSetValueWarnings(t *testing.T) {
	var c testConfig

	data := localData{userStruct: reflect.ValueOf(&c).Elem()}

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

	for _, test := range tests {
		data.setValue(test.Field, test.SetTo)
		if data.messages == nil {
			t.Errorf("TestSetValue failed for field '%s'", test.Field)
		}
	}
}

func TestSetStruct(t *testing.T) {
	testStruct := struct {
		Field struct {
			Test bool
		} `cmd:"struct"`
	}{}
	data := localData{userStruct: reflect.ValueOf(&testStruct).Elem()}
	data.setValue("Field", `{"test":false}`)
	if data.messages == nil {
		t.Fail()
	}
}
