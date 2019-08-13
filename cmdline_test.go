package igconfig

import (
	"reflect"
	"testing"
)

func TestCmdlineValues(t *testing.T) {
	const funcName = "TestCmdlineValues"

	args := []string{"-t", "-n", "Piet", "--port", "1234", "--hostname=bol.com", "--age", "25", "--salary", "1500.00"}

	var c testConfig
	data := localData{userStruct: reflect.ValueOf(&c).Elem()}

	err := data.loadCmdline(args)
	if err != nil {
		t.Errorf("%s failed: %s", funcName, err.Error())
	}

	if c.Name != "Piet" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Piet")
	}
	if c.Age != 25 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 25)
	}
	if c.Salary != 1500.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 1500.0)
	}
	if c.Host != "bol.com" {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, "bol.com")
	}
	if c.Port != 1234 {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, 1234)
	}
	if c.Secure != true {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, true)
	}
}
