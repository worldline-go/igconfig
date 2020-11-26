package loader

import (
	"reflect"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/internal"
)

var _ Loader = Default{}

const DefaultTag = "default"

type Default struct{}

// Load loads the config struct fields with their default value as defined in the tags.
func (d Default) Load(_ string, to interface{}) error {
	refVal, err := internal.GetReflectElem(to)
	if err != nil {
		return err
	}

	return d.ReflectLoad("", refVal)
}

func (d Default) ReflectLoad(_ string, to reflect.Value) error {
	t := to.Type()

	for i := 0; i < t.NumField(); i++ {
		toField := to.Field(i)
		if (toField.IsValid() && !toField.IsZero()) || !toField.CanSet() {
			// Value is already set or not settable at all, skip it.
			continue
		}

		typeField := t.Field(i)

		// If it is a struct - try to set it's inner fields to.
		if typeField.Type.Kind() == reflect.Struct {
			if err := d.ReflectLoad("", toField); err != nil {
				return err
			}

			continue
		}

		if v, ok := typeField.Tag.Lookup(DefaultTag); ok {
			if err := internal.SetStructFieldValue(typeField.Name, v, to); err != nil {
				return err
			}
		}
	}

	return nil
}
