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

			v := internal.TagValueByKeys(f.Tag, DefaultTag)
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
