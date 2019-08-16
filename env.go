package igconfig

import (
	"os"
	"strings"
)

// loadEnv loads config values from the environment
func (m *localData) loadEnv() {
	t := m.userStruct.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tags := field.Tag.Get("env")
		if tags == "" {
			m.testEnv(field.Name, field.Name)
			continue
		}
		nn := strings.Split(tags, ",")
		for _, n := range nn {
			if m.testEnv(field.Name, n) {
				break
			}
		}
	}
}

// testEnv tests for an environment variable, and if found sets the field's value
func (m *localData) testEnv(fieldName, n string) bool {
	if v, ok := os.LookupEnv(n); ok {
		m.setValue(fieldName, v)
		return true
	}
	if v, ok := os.LookupEnv(strings.ToUpper(n)); ok {
		m.setValue(fieldName, v)
		return true
	}
	if v, ok := os.LookupEnv(strings.ToLower(n)); ok {
		m.setValue(fieldName, v)
		return true
	}
	return false
}
