package igconfig

import (
	"fmt"
	"reflect"
)

// loadDefaults loads the config struct fields with their default value as defined in the tags
func (m *myConfig) loadDefaults() error {
	const funcName = "loadDefaults"

	v := reflect.ValueOf(m.c)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("%s: input parameter not a pointer", funcName)
	}

	t := v.Elem().Type()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("%s: input parameter not a struct", funcName)
	}

	for i := 0; i < t.NumField(); i++ {
		m.f = t.Field(i)
		if v, ok := m.f.Tag.Lookup("default"); ok {
			m.setValue(v)
		}
	}

	return nil
}
