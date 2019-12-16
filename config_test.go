package igconfig

import (
	"io/ioutil"
	"os"
	"testing"
)

type testConfig struct {
	Name    string  `cfg:"settle_name"    env:"name"           cmd:"name,n"           default:"Jan"`
	Age     uint    `cfg:"age"            env:"age"            cmd:"age,a"            default:"18"`
	Salary  float64 `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"2000.00"  loggable:"false"`
	Host    string  `cfg:"host,hostname"  env:"host,hostname"  cmd:"host,hostname,h"  default:"localhost"`
	Address string  `cfg:"ADDRESS"        env:"ADDRESS"        default:"localhost"`
	Port    int     `cfg:"port"           env:"port"           cmd:"port,p"           default:"8080"`
	Secure  bool    `cfg:"secure,ssl,tls" env:"secure,ssl,tls" cmd:"secure,ssl,tls,t" default:"false"    loggable:"false"`
	Unused  []string
}

type badDefaults struct {
	Age    uint    `cfg:"age"            env:"age"            cmd:"age,a"            default:"haha"`
	Salary float64 `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"haha"`
	Port   int     `cfg:"port"           env:"port"           cmd:"port,p"           default:"haha"`
}

func TestNewLocalData(t *testing.T) {
	i := 0

	if _, err := newLocalData(i); err == nil {
		t.Error("failed to test for invalid input parameter (not pointer)")
	}
	if _, err := newLocalData(&i); err == nil {
		t.Error("failed to test for invalid input parameter (not struct)")
	}

	if err := LoadConfigDefaults(i); err == nil {
		t.Error("failed to test LoadConfigDefaults")
	}
	if err := LoadConfigFile(i, "haha"); err == nil {
		t.Error("failed to test LoadConfigFile")
	}
	if err := LoadConfigEnv(i); err == nil {
		t.Error("failed to test LoadConfigEnv")
	}
	if err := LoadConfigCmdline(i); err == nil {
		t.Error("failed to test LoadConfigCmdline")
	}
	if err := LoadConfig(i, "haha", true, true); err == nil {
		t.Error("failed to test LoadConfig")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	c1 := badDefaults{}
	if LoadConfigDefaults(&c1) == nil {
		t.Error("failed to test for bad defaults")
	}

	c2 := testConfig{}
	if LoadConfigDefaults(&c2) != nil {
		t.Error("failed to load good defaults")
	}
}

func TestLoadConfigFile(t *testing.T) {
	var c testConfig
	if LoadConfigFile(&c, "/this/is/not/a/file") == nil {
		t.Error("failed to test for invalid file name")
	}

	if LoadConfigFile(&c, "/dev/null") != nil {
		t.Error("failed to load fileName /dev/null")
	}

	const fileName = "/tmp/TestFileBadData.cfg"
	const fileData = "age=haha"

	if err := ioutil.WriteFile(fileName, []byte(fileData), 0644); err != nil {
		t.Errorf("could not write temporary fileName '%s'", fileName)
		return
	}
	defer os.Remove(fileName)

	if LoadConfigFile(&c, fileName) == nil {
		t.Error("failed to check for parsing errors")
	}
}

func TestLoadConfigEnv(t *testing.T) {
	var c testConfig
	if LoadConfigEnv(&c) != nil {
		t.Error("failed to load environment")
	}

	if os.Setenv("port", "haha") != nil {
		t.Error("could not set environment variable 'Port'")
	}

	if LoadConfigEnv(&c) == nil {
		t.Error("failed to test for parsing error")
	}

	if os.Unsetenv("port") != nil {
		t.Error("could not unset environment variable 'Port'")
	}
}

func TestLoadConfigCmdline(t *testing.T) {
	var c testConfig

	os.Args = []string{"program"}
	if LoadConfigCmdline(&c) != nil {
		t.Error("failed to load command-line parameters")
	}

	os.Args = []string{"program", "--age", "haha"}
	if LoadConfigCmdline(&c) == nil {
		t.Error("failed to test for parsing error")
	}

	os.Args = []string{"program"}
}

func TestLoadConfig(t *testing.T) {
	c1 := badDefaults{}
	if LoadConfig(&c1, "haha", true, true) == nil {
		t.Error("failed to check for bad defaults")
	}

	var c2 testConfig
	if LoadConfig(&c2, "/this/is/not/a/file", true, true) == nil {
		t.Error("failed to check for bad fileName")
	}

	if LoadConfig(&c2, "/dev/null", true, true) != nil {
		t.Error("failed to load configuration")
	}

	os.Args = []string{"program", "--age", "haha", "--salary", "nothing"}

	c2 = testConfig{}
	if LoadConfig(&c2, "", true, true) == nil {
		t.Error("failed to check for parsing errors")
	}

	os.Args = []string{"program"}
}
