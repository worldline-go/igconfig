package igconfig

import (
	"testing"
)

func TestDefaultsValues(t *testing.T) {
	const funcName = "TestDefaultsValues"
	const host = "test"
	const port = 7788

	var c testConfig = testConfig{
		Port: port,
		Host: host,
	}
	data, err := newLocalData(&c)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", funcName, err.Error())
	}

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
	if c.Host != host {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, host)
	}
	if c.Port != port {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, port)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
}
