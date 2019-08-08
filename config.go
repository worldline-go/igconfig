package igconfig

import (
	"fmt"
	"reflect"
	"strings"
)

// myConfig contains local variables during config processing
type myConfig struct {
	c        interface{}
	file     string
	warnings []string
	f        reflect.StructField
}

// LoadConfigDefaults loads a configuration struct with its defaults.
func LoadConfigDefaults(c interface{}) error {
	data := myConfig{c: c, warnings: make([]string, 0)}
	if err := data.loadDefaults(); err != nil {
		return err
	}

	if len(data.warnings) > 0 {
		return fmt.Errorf("LoadConfigDefaults warnings found:\n%s", strings.Join(data.warnings, "\n"))
	}

	return nil
}

// LoadConfigFile loads a configuration struct from a file.
func LoadConfigFile(c interface{}, file string) error {
	data := myConfig{c: c, file: file, warnings: make([]string, 0)}

	if err := data.loadFile(); err != nil {
		return err
	}

	if len(data.warnings) > 0 {
		return fmt.Errorf("LoadConfigFile warnings found:\n%s", strings.Join(data.warnings, "\n"))
	}

	return nil
}

// LoadConfigEnv loads a configuration struct from the environment.
func LoadConfigEnv(c interface{}) error {
	data := myConfig{c: c, warnings: make([]string, 0)}

	data.loadEnv()

	if len(data.warnings) > 0 {
		return fmt.Errorf("LoadConfigEnv warnings found:\n%s", strings.Join(data.warnings, "\n"))
	}

	return nil
}

// LoadConfigCmdline loads a configuration struct from the command-line parameters.
func LoadConfigCmdline(c interface{}) error {
	data := myConfig{c: c, warnings: make([]string, 0)}

	data.loadCmdline()

	if len(data.warnings) > 0 {
		return fmt.Errorf("LoadConfigCmdline warnings found:\n%s", strings.Join(data.warnings, "\n"))
	}

	return nil
}

// LoadConfig loads a configuration struct from a file, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
// This is a convenience function encapsulating all individual functions above.
func LoadConfig(c interface{}, file string, env, cmd bool) error {
	data := myConfig{c: c, file: file, warnings: make([]string, 0)}

	err := data.loadDefaults()
	if err != nil {
		return err
	}

	if file != "" {
		err = data.loadFile()
	}

	if env {
		data.loadEnv()
	}

	if cmd {
		data.loadCmdline()
	}

	if err != nil {
		return err
	}

	if len(data.warnings) > 0 {
		return fmt.Errorf("LoadConfig warnings found:\n%s", strings.Join(data.warnings, "\n"))
	}

	return nil
}
