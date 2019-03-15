package igconfig

import (
	"reflect"
	"errors"
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
)

// loadDefaults loads the config struct fields with their default value as defined in the tags
func loadDefaults(c interface{}) error {
	iglog.DebugFull("LoadConfig: loading defaults")

	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr {
		return errors.New("LoadConfig: input parameter not a pointer")
	}
	
	t := v.Elem().Type()
	if t.Kind() != reflect.Struct {
		return errors.New("LoadConfig: input parameter not a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if v, ok := f.Tag.Lookup("default"); ok {
			setValue(c, f, v)
		}
	}

	return nil
}
