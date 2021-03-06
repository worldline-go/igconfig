package loader

import (
	"context"
	"reflect"
	"strings"

	"github.com/worldline-go/igconfig/internal"
)

var _ Loader = Default{}

// DefaultTag is a tag name for default value.
const DefaultTag = "default"

// Default is a loader that loads config struct fields with their default value as defined in the tags.
type Default struct{}

// LoadWithContext loads the config struct fields with their default value as defined in the tags.
func (l Default) LoadWithContext(_ context.Context, _ string, to interface{}) error {
	it := internal.StructIterator{
		Value:         to,
		NoUpdate:      true,
		FieldNameFunc: l.FieldNameFunc,
		IteratorFunc:  l.IteratorFunc,
	}

	return it.Iterate()
}

// Load is just same as LoadWithContext without context.
func (l Default) Load(_ string, to interface{}) error {
	return l.LoadWithContext(context.TODO(), "", to)
}

// FieldNameFunc defined field retrieval function for Default loader.
//
// This function will return string in format of <field_name>:<default_value>
// or just "-" if no default value is defined.
func (l Default) FieldNameFunc(outer string, field reflect.StructField) string {
	isStruct := internal.IsStruct(field.Type)

	v := internal.TagValueByKeys(field.Tag, DefaultTag)
	if internal.IsTagOmitted(v) && !isStruct { // If no default value and is not struct - skip such field.
		return "-"
	}

	// This is situation 'default:"-"'. For structs specifically.
	// If this is the case - then whole struct will be skipped.
	if internal.IsTagSkip(v) {
		return "-"
	}

	fieldName := internal.PlainFieldNameWithPath(outer, field)
	if isStruct {
		return fieldName
	}

	return fieldName + ":" + strings.Join(v, ",")
}

// IteratorFunc returns a setter function for setting fields.
func (l Default) IteratorFunc(fieldName string, field reflect.Value) error {
	sl := strings.SplitN(fieldName, ":", 2)

	return internal.SetReflectValueString(sl[0], sl[1], field)
}
