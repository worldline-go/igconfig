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
		if to.Field(i).IsValid() && !to.Field(i).IsZero() {
			// Value is already set, skip it.
			continue
		}

		field := t.Field(i)
		if v, ok := field.Tag.Lookup(DefaultTag); ok {
			if err := internal.SetStructFieldValue(field.Name, v, to); err != nil {
				return err
			}
		}
	}

	return nil
}
