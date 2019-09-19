package igconfig

import (
	"os"
	"strings"
	"testing"
)

func TestEnvValues(t *testing.T) {
	const funcName = "TestEnvValues"

	if err := os.Setenv("HOSTNAME", "127.0.0.1"); err != nil {
		t.Errorf("%s could not set environment variable 'HOSTNAME'", funcName)
	}
	if err := os.Setenv("port", "12345"); err != nil {
		t.Errorf("%s could not set environment variable 'Port'", funcName)
	}
	if err := os.Setenv("age", "44"); err != nil {
		t.Errorf("%s could not set environment variable 'age'", funcName)
	}
	if err := os.Setenv("address", "should_not_be_set"); err != nil {
		t.Errorf("%s could not set environment variable 'address'", funcName)
	}

	var c testConfig
	data, err := newLocalData(&c)
	if err != nil {
		t.Fatalf("%s: should not fail: %s", funcName, err.Error())
	}
	data.loadDefaults()

	data.loadEnv()
	if len(data.messages) != 0 {
		t.Errorf("%s failed\n%s", funcName, strings.Join(data.messages, "\n"))
	}

	if c.Name != "Jan" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Jan")
	}
	if c.Age != 44 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 44)
	}
	if c.Salary != 2000.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 2000.0)
	}
	if c.Host != "127.0.0.1" {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, "127.0.0.1")
	}
	if c.Port != 12345 {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, 12345)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
}
