package igconfig

import (
	"reflect"
	"testing"
)

func TestDefaultsValues(t *testing.T) {
	const funcName = "TestDefaultsValues"
	const localhost = "localhost"

	var c testConfig
	data := localData{userStruct: reflect.ValueOf(&c).Elem()}

	data.loadDefaults()

	if c.Name != "Jan" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Jan")
	}
	if c.Age != 18 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 18)
	}
	if c.Salary != 2000.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 2000.0)
	}
	if c.Host != localhost {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, localhost)
	}
	if c.Port != 8080 {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, 8080)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
}
