package igconfig

import (
	"errors"
	"reflect"
)

// loadDefaults loads the config struct fields with their default value as defined in the tags
func (m *myConfig) loadDefaults() error {
	v := reflect.ValueOf(m.c)
	if v.Kind() != reflect.Ptr {
		return errors.New("LoadConfig: input parameter not a pointer")
	}

	t := v.Elem().Type()
	if t.Kind() != reflect.Struct {
		return errors.New("LoadConfig: input parameter not a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		m.f = t.Field(i)
		if v, ok := m.f.Tag.Lookup("default"); ok {
			m.setValue(v)
		}
	}

	return nil
}
