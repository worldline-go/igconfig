package igconfig

import (
	"os"
	"reflect"
	"strings"
)

// loadEnv loads config values from the environment
func (m *localData) loadEnv() {
	v := reflect.ValueOf(m.userStruct)
	e := v.Elem()
	t := e.Type()

	for i := 0; i < t.NumField(); i++ {
		m.fld = t.Field(i)
		if !m.testEnv(m.fld.Name) {
			nn := strings.Split(m.fld.Tag.Get("env"), ",")
			for _, n := range nn {
				if m.testEnv(n) {
					break
				}
			}
		}
	}
}

// testEnv tests for an environment variable, and if found sets the field's value
func (m *localData) testEnv(n string) bool {
	if v, ok := os.LookupEnv(n); ok {
		m.setValue(v)
		return true
	}
	if v, ok := os.LookupEnv(strings.ToUpper(n)); ok {
		m.setValue(v)
		return true
	}
	if v, ok := os.LookupEnv(strings.ToLower(n)); ok {
		m.setValue(v)
		return true
	}
	return false
}
