package igconfig

import (
	"testing"
	"os"
)

type MyConfig struct {
	Name   string  `cfg:"settle_name" env:"name" cmd:"name,n" default:"Jan"`
	Age    int     `cfg:"age" env:"age" cmd:"age,a" default:"18"`
	Salary float64 `cfg:"salary" env:"salary" cmd:"salary,s" default:"2000.00"`
	Host   string  `cfg:"host,hostname" env:"host,hostname" cmd:"host,hostname,h" default:"localhost"`
	Port   int     `cfg:"port" env:"port" cmd:"port,p" default:"8080"`
	Secure bool    `cfg:"secure,ssl,tls" env:"secure,ssl,tls" cmd:"secure,ssl,tls,t" default:"false"`
}

func TestDefaults(t *testing.T) {
	var c MyConfig

	err := LoadConfig(&c, "", false, false, false)
	if err != nil {
		t.Errorf("TestDefaults failed: %s", err.Error())
	}

	if c.Name != "Jan" {
		t.Errorf("TestEnv name mismatch; got: %s; want: %s", c.Name, "Jan")
	}
	if c.Age != 18 {
		t.Errorf("TestEnv age mismatch; got: %d; want: %d", c.Age, 18)
	}
	if c.Salary != 2000.0 {
		t.Errorf("TestEnv salary mismatch; got: %.2f; want: %.2f", c.Salary, 2000.0)
	}
	if c.Host != "localhost" {
		t.Errorf("TestEnv host mismatch; got: %s; want: %s", c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("TestEnv port mismatch; got: %d; want: %d", c.Port, 8080)
	}
	if c.Secure != false {
		t.Errorf("TestEnv secure mismatch; got: %t; want: %t", c.Secure, false)
	}
}

func TestEnv(t *testing.T) {
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("Port", "12345")
	os.Setenv("age", "44")
	
	var c MyConfig

	err := LoadConfig(&c, "", true, false, false)
	if err != nil {
		t.Errorf("TestEnv failed: %s", err.Error())
	}

	if c.Name != "Jan" {
		t.Errorf("TestEnv name mismatch; got: %s; want: %s", c.Name, "Jan")
	}
	if c.Age != 44 {
		t.Errorf("TestEnv age mismatch; got: %d; want: %d", c.Age, 44)
	}
	if c.Salary != 2000.0 {
		t.Errorf("TestEnv salary mismatch; got: %.2f; want: %.2f", c.Salary, 2000.0)
	}
	if c.Host != "127.0.0.1" {
		t.Errorf("TestEnv host mismatch; got: %s; want: %s", c.Host, "127.0.0.1")
	}
	if c.Port != 12345 {
		t.Errorf("TestEnv port mismatch; got: %d; want: %d", c.Port, 12345)
	}
	if c.Secure != false {
		t.Errorf("TestEnv secure mismatch; got: %t; want: %t", c.Secure, false)
	}
}

func TestCmdline(t *testing.T) {
	os.Args = []string {"program", "-t", "-n", "Piet", "--port", "1234", "--hostname=bol.com"}
	
	var c MyConfig

	err := LoadConfig(&c, "", false, true, false)
	if err != nil {
		t.Errorf("TestCmdline failed: %s", err.Error())
	}

	if c.Name != "Piet" {
		t.Errorf("TestEnv name mismatch; got: %s; want: %s", c.Name, "Piet")
	}
	if c.Age != 18 {
		t.Errorf("TestEnv age mismatch; got: %d; want: %d", c.Age, 18)
	}
	if c.Salary != 2000.0 {
		t.Errorf("TestEnv salary mismatch; got: %.2f; want: %.2f", c.Salary, 2000.0)
	}
	if c.Host != "bol.com" {
		t.Errorf("TestEnv host mismatch; got: %s; want: %s", c.Host, "bol.com")
	}
	if c.Port != 1234 {
		t.Errorf("TestEnv port mismatch; got: %d; want: %d", c.Port, 1234)
	}
	if c.Secure != true {
		t.Errorf("TestEnv secure mismatch; got: %t; want: %t", c.Secure, true)
	}
}
