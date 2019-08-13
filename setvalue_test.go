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

	data := localData{userStruct: &c}

	v := reflect.ValueOf(data.userStruct)
	e := v.Elem()
	tp := e.Type()

	data.fld, _ = tp.FieldByName("Age")
	data.setValue("haha")
	if data.messages == nil {
		t.Errorf("TestSetValue failed to test for uint parsing error")
	} else {
		data.messages = nil
	}

	data.fld, _ = tp.FieldByName("Port")
	data.setValue("haha")
	if data.messages == nil {
		t.Errorf("TestSetValue failed to test for int parsing error")
	} else {
		data.messages = nil
	}

	data.fld, _ = tp.FieldByName("Salary")
	data.setValue("haha")
	if data.messages == nil {
		t.Errorf("TestSetValue failed to test for float parsing error")
	} else {
		data.messages = nil
	}

	data.fld, _ = tp.FieldByName("Unused")
	data.setValue("1.0")
	if data.messages == nil {
		t.Errorf("TestSetValue failed to test for unsupported type")
	}
}
