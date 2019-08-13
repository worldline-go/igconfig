package igconfig

import (
	"reflect"
)

// loadDefaults loads the config struct fields with their default value as defined in the tags
func (m *localData) loadDefaults() {
	v := reflect.ValueOf(m.userStruct)
	t := v.Elem().Type()

	for i := 0; i < t.NumField(); i++ {
		m.fld = t.Field(i)
		if v, ok := m.fld.Tag.Lookup("default"); ok {
			m.setValue(v)
		}
	}
}
