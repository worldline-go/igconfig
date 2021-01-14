package loader

import (
	"reflect"
	"strings"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/internal"
)

var _ Loader = Default{}

const DefaultTag = "default"

type Default struct{}

// Load loads the config struct fields with their default value as defined in the tags.
func (d Default) Load(_ string, to interface{}) error {
	it := internal.StructIterator{
		Value:    to,
		NoUpdate: true,
		// This function will return string in format of <field_name>:<default_value>
		// or just "-" if no default value is defined.
		FieldNameFunc: func(outer string, f reflect.StructField) string {
			isStruct := internal.IsStruct(f.Type)

			v := internal.TagValueByKeys(f, DefaultTag)
			if internal.IsTagOmitted(v) && !isStruct { // If no default value and is not struct - skip such field.
				return "-"
			}

			if internal.IsTagSkip(v) { // This is situation 'default:"-"'. For structs specifically.
				return "-"
			}

			fieldName := internal.PlainFieldNameWithPath(outer, f)
			if isStruct {
				return fieldName
			}

			return fieldName + ":" + v[0]
		},
		IteratorFunc: func(fieldName string, field reflect.Value) error {
			sl := strings.SplitN(fieldName, ":", 2)

			return internal.SetReflectValueString(sl[0], sl[1], field)
		},
	}

	return it.Iterate()
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

	}

	return nil
}
