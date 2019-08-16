package igconfig

import (
	"bytes"
	"testing"
)

func TestFileBadData(t *testing.T) {
	var (
		buf      bytes.Buffer
		funcName = "TestFileBadData"
	)
	tests := []struct {
		FileData   string
		ShouldFail bool
	}{
		{
			FileData:   "age=haha",
			ShouldFail: true,
		},
		{
			FileData:   "age",
			ShouldFail: false,
		},
	}

	var c testConfig
	data, err := newLocalData(&c)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", funcName, err.Error())
	}

	for _, test := range tests {
		buf.WriteString(test.FileData)
		if data.loadReader(&buf) != nil {
			t.Errorf("%s failed to read configuration", funcName)
		}
		if test.ShouldFail && len(data.messages) == 0 {
			t.Errorf("%s failed to check for parsing error", funcName)
		}
		data.messages = nil
		buf.Reset()
	}
}

func TestFileOverwriteDefault(t *testing.T) {
	var (
		buf      bytes.Buffer
		funcName = "TestFileOverwriteDefault"
		fileData = "settle_name="
	)
	buf.WriteString(fileData)

	var c testConfig
	data, err := newLocalData(&c)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", funcName, err.Error())
	}
	data.loadDefaults()

	if data.loadReader(&buf) != nil {
		t.Errorf("%s failed to read configuration", funcName)
	}
	if len(data.messages) != 0 {
		t.Errorf("%s got parsing error(s)", funcName)
	}

	if c.Name != "" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "")
	}
}

func TestFileSimple(t *testing.T) {
	var (
		buf      bytes.Buffer
		funcName = "TestFileSimple"
		fileData = "age=28\nsalary=1800.00\nsettle_name=Jantje"
	)
	buf.WriteString(fileData)

	var c testConfig
	data, err := newLocalData(&c)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", funcName, err.Error())
	}
	data.loadDefaults()

	if data.loadReader(&buf) != nil {
		t.Errorf("%s could not load config", funcName)
	}
	if len(data.messages) != 0 {
		t.Errorf("%s got parsing error(s)", funcName)
	}

	if c.Name != "Jantje" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Jantje")
	}
	if c.Age != 28 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 28)
	}
	if c.Salary != 1800.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 1800.0)
	}
	if c.Host != "localhost" {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, 8080)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
}

func TestFileComplex(t *testing.T) {
	var (
		buf      bytes.Buffer
		funcName = "TestFileComplex"
		fileData = "// Age\nage=28\n#Salary\nsalary=1800.00\n\nsettle_name=Jantje\n ## Name of subject ##\nwrong=test\n\n"
	)
	buf.WriteString(fileData)

	var c testConfig
	data, err := newLocalData(&c)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", funcName, err.Error())
	}
	data.loadDefaults()

	if data.loadReader(&buf) != nil {
		t.Errorf("%s could not load config", funcName)
	}
	if len(data.messages) != 0 {
		t.Errorf("%s got parsing error(s)", funcName)
	}

	if c.Name != "Jantje" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Jantje")
	}
	if c.Age != 28 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 28)
	}
	if c.Salary != 1800.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 1800.0)
	}
	if c.Host != "localhost" {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, 8080)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
}
