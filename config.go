package igconfig

import (
	"fmt"
	"reflect"
	"strings"
)

// Printer interface can be supplied by a calling process to print the configuration after it has been loaded
type Printer interface {
	PrintConfig(c interface{})
}

// myConfig contains local variables during config processing
type myConfig struct {
	c        interface{}
	file     string
	warnings []string
	f        reflect.StructField
}

// LoadConfig loads configuration parameters from a file, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
func LoadConfig(c interface{}, file string, env, cmd bool) error {
	data := myConfig{c: c, file: file, warnings: make([]string, 0)}

	err := data.loadDefaults()
	if err != nil {
		return err
	}

	if file != "" {
		err = data.loadConfigFile()
	}

	if env {
		data.loadEnv()
	}

	if cmd {
		data.loadCmdline()
	}

	if len(data.warnings) > 0 {
		return fmt.Errorf("LoadConfig warnings found:\n%s", strings.Join(data.warnings, "\n"))
	}

	return err
}

// LoadConfigAndPrint loads configuration parameters from a file, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
// It will use the supplied Printer interface to allow the caller to print the configuration
func LoadConfigAndPrint(c interface{}, file string, env, cmd bool, p Printer) error {
	err := LoadConfig(c, file, env, cmd)
	if err != nil {
		return err
	}

	if p != nil {
		p.PrintConfig(c)
	}

	return nil
}
