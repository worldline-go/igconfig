package igconfig

import (
	"io/ioutil"
	"os"
	"testing"
)

type testConfig struct {
	Name   string  `cfg:"settle_name"    env:"name"           cmd:"name,n"           default:"Jan"`
	Age    uint    `cfg:"age"            env:"age"            cmd:"age,a"            default:"18"`
	Salary float64 `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"2000.00"  loggable:"false"`
	Host   string  `cfg:"host,hostname"  env:"host,hostname"  cmd:"host,hostname,h"  default:"localhost"`
	Port   int     `cfg:"port"           env:"port"           cmd:"port,p"           default:"8080"`
	Secure bool    `cfg:"secure,ssl,tls" env:"secure,ssl,tls" cmd:"secure,ssl,tls,t" default:"false"    loggable:"false"`
	Unused []string
}

type badDefaults struct {
	Age    uint    `cfg:"age"            env:"age"            cmd:"age,a"            default:"haha"`
	Salary float64 `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"haha"`
	Port   int     `cfg:"port"           env:"port"           cmd:"port,p"           default:"haha"`
}

func TestConfigParm(t *testing.T) {
	const funcName = "TestConfigParm"

	i := 0

	if testConfigParm(i) == nil {
		t.Errorf("%s failed to test for invalid input parameter (not pointer)", funcName)
	}

	if testConfigParm(&i) == nil {
		t.Errorf("%s failed to test for invalid input parameter (not struct)", funcName)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	const funcName = "TestLoadConfigDefaults"

	c1 := badDefaults{}
	if LoadConfigDefaults(&c1) == nil {
		t.Errorf("%s failed to test for bad defaults", funcName)
	}

	c2 := testConfig{}
	if LoadConfigDefaults(&c2) != nil {
		t.Errorf("%s failed to load good defaults", funcName)
	}
}

func TestLoadConfigFile(t *testing.T) {
	const funcName = "TestLoadConfigFile"

	i := 0
	if LoadConfigFile(i, "/tmp/haha.txt") == nil {
		t.Errorf("%s failed to test for invalid fileName name", funcName)
	}

	var c testConfig
	if LoadConfigFile(&c, "/dev/null") != nil {
		t.Errorf("%s failed to load fileName /dev/null", funcName)
	}

	const fileName = "/tmp/TestFileBadData.cfg"
	const fileData = "age=haha"

	if err := ioutil.WriteFile(fileName, []byte(fileData), 0644); err != nil {
		t.Errorf("%s could not write temporary fileName '%s'", funcName, fileName)
		return
	}
	defer os.Remove(fileName)

	if LoadConfigFile(&c, fileName) == nil {
		t.Errorf("%s failed to check for parsing errors", funcName)
	}
}

func TestLoadConfigEnv(t *testing.T) {
	const funcName = "TestLoadConfigEnv"

	var c testConfig
	if LoadConfigEnv(&c) != nil {
		t.Errorf("%s failed to load environment", funcName)
	}

	if os.Setenv("Port", "haha") != nil {
		t.Errorf("%s could not set environment variable 'Port'", funcName)
	}

	if LoadConfigEnv(&c) == nil {
		t.Errorf("%s failed to test for parsing error", funcName)
	}

	if os.Unsetenv("Port") != nil {
		t.Errorf("%s could not unset environment variable 'Port'", funcName)
	}
}

func TestLoadConfigCmdline(t *testing.T) {
	const funcName = "TestLoadConfigCmdline"

	var c testConfig

	os.Args = []string{"program"}
	if LoadConfigCmdline(&c) != nil {
		t.Errorf("%s failed to load command-line parameters", funcName)
	}

	os.Args = []string{"program", "--age", "haha"}
	if LoadConfigCmdline(&c) == nil {
		t.Errorf("%s failed to test for parsing error", funcName)
	}

	os.Args = []string{"program"}
}

func TestLoadConfig(t *testing.T) {
	const funcName = "TestLoadConfig"

	var c testConfig
	if LoadConfig(&c, "/tmp/haha.txt", true, true) == nil {
		t.Errorf("%s failed to check for bad fileName", funcName)
	}

	if LoadConfig(&c, "/dev/null", true, true) != nil {
		t.Errorf("%s failed to load configuration", funcName)
	}

	os.Args = []string{"program", "--age", "haha", "--salary", "nothing"}

	c = testConfig{}
	if LoadConfig(&c, "", true, true) == nil {
		t.Errorf("%s failed to check for parsing errors", funcName)
	}

	os.Args = []string{"program"}
}
