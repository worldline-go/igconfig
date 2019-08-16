package igconfig

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// localData contains local variables during config processing
type localData struct {
	userStruct reflect.Value
	messages   []string
}

// newLocalData tests if the supplied parameter is a valid pointer to a struct and returns localData
func newLocalData(c interface{}) (localData, error) {
	data := localData{}
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr {
		return data, errors.New("input parameter not a pointer")
	}

	if v.Elem().Type().Kind() != reflect.Struct {
		return data, errors.New("input parameter not a struct")
	}
	data.userStruct = v.Elem()
	return data, nil
}

// checkWarnings tests if messages were found and returns with an appropriate error
func (m *localData) checkWarnings(funcName string) error {
	if len(m.messages) > 0 {
		return fmt.Errorf("%s errors found:\n%s", funcName, strings.Join(m.messages, "\n"))
	}
	return nil
}

// LoadConfigDefaults loads a configuration struct with its defaults.
func LoadConfigDefaults(c interface{}) error {
	const funcName = "LoadConfigDefaults"

	data, err := newLocalData(c)
	if err != nil {
		return fmt.Errorf("%s %s", funcName, err.Error())
	}

	data.loadDefaults()
	return data.checkWarnings(funcName)
}

// LoadConfigFile loads a configuration struct from a fileName.
func LoadConfigFile(c interface{}, file string) error {
	const funcName = "LoadConfigFile"

	data, err := newLocalData(c)
	if err != nil {
		return fmt.Errorf("%s %s", funcName, err.Error())
	}
	if err := data.loadFile(file); err != nil {
		return err
	}

	return data.checkWarnings(funcName)
}

// LoadConfigEnv loads a configuration struct from the environment.
func LoadConfigEnv(c interface{}) error {
	const funcName = "LoadConfigEnv"

	data, err := newLocalData(c)
	if err != nil {
		return fmt.Errorf("%s %s", funcName, err.Error())
	}
	data.loadEnv()
	return data.checkWarnings(funcName)
}

// LoadConfigCmdline loads a configuration struct from the command-line parameters.
func LoadConfigCmdline(c interface{}) error {
	const funcName = "LoadConfigCmdline"

	data, err := newLocalData(c)
	if err != nil {
		return fmt.Errorf("%s %s", funcName, err.Error())
	}
	if err := data.loadCmdline(os.Args[1:]); err != nil {
		return err
	}

	return data.checkWarnings(funcName)
}

// LoadConfig loads a configuration struct from a fileName, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
// This is a convenience function encapsulating all individual functions above.
func LoadConfig(c interface{}, file string, env, cmd bool) error {
	const funcName = "LoadConfig"

	data, err := newLocalData(c)
	if err != nil {
		return fmt.Errorf("%s %s", funcName, err.Error())
	}
	// if defaults fail there's an error in the struct so we return immediately
	data.loadDefaults()
	if data.messages != nil {
		return data.checkWarnings(funcName)
	}

	// all other "user" inputs should be checked and all messages reported, so the user can fix everything and
	// not have to keep trying after the first fail
	if file != "" {
		err = data.loadFile(file)
	}

	if env {
		data.loadEnv()
	}

	if cmd {
		if e := data.loadCmdline(os.Args[1:]); e != nil && err == nil {
			err = e
		}
	}

	if err != nil {
		return err
	}

	return data.checkWarnings(funcName)
}
