package internal

import (
	"errors"
	"os"
	"reflect"
	"strings"
)

//nolint:golint
const DefaultConfigTag = "cfg"

//nolint:golint
const SkipFieldTagValue = "-"

//nolint:golint
var ErrInputIsNotPointerOrStruct = errors.New("input value is not struct or not pointer")

//nolint:golint
var ErrNoIterValue = errors.New("no value to iterate")

//nolint:golint
type FieldNameFunc func(outerName string, currentField reflect.StructField) string

//nolint:golint
type IteratorFunc func(fieldName string, field reflect.Value) error

// StructIterator allows to traverse structures and call predefined function on each field.
//
// All loaders that use internal.StructIterator will not be able to use multiple names of fields!
type StructIterator struct {
	Value        interface{}
	ReflectValue reflect.Value
	BaseName     string
	// NoUpdate will specify that IteratorFunc cannot be run on fields with non-zero values.
	NoUpdate      bool
	FieldNameFunc FieldNameFunc
	IteratorFunc  IteratorFunc
}

// Iterate will iterate over every field and will execute FieldNameFunc and IteratorFunc on them.
//
// It will dive into all inner structs and will also iterate fields there.
//
// Pointer fields that are equal to nil will be initialized.
//
// Fields(but not structs) can implement encoding.TextUnmarshaler to be able to set values with custom logic.
func (it StructIterator) Iterate() error {
	if err := it.ValidateData(); err != nil {
		return err
	}

	val := it.ReflectValue.Elem()
	valType := val.Type()

	for i := 0; i < valType.NumField(); i++ {
		toField, structField := val.Field(i), valType.Field(i)
		// Do not process unexported fields.
		// PkgPath is set only for unexported fields: https://golang.org/pkg/reflect/#StructField
		if structField.PkgPath != "" {
			continue
		}

		fieldName := it.FieldNameFunc(it.BaseName, structField)

		if ShouldSkipField(toField, fieldName, it.NoUpdate) {
			continue
		}

		// Set zero-value to valid pointer values when they are being processed.
		if toField.Kind() == reflect.Ptr && toField.IsNil() {
			toField.Set(reflect.New(toField.Type().Elem()))

			toField = toField.Elem()
		}

		// If it is a struct - try to set it's inner fields to.
		// TODO: have a list of non-struct types for types like sql.Null* and null.*
		if toField.Kind() == reflect.Struct && toField.Type() != TimeType {
			subIter := StructIterator{
				ReflectValue:  toField.Addr(), // This is just simple solution for using pointer structs as inputs.
				BaseName:      fieldName,
				FieldNameFunc: it.FieldNameFunc,
				IteratorFunc:  it.IteratorFunc,
			}
			// Do the same iteration on inner struct.
			if err := subIter.Iterate(); err != nil {
				return err
			}

			continue
		}

		if err := it.IteratorFunc(fieldName, toField); err != nil {
			return err
		}
	}

	return nil
}

// ValidateData makes sure that input data in iterator will be able to use it.
func (it *StructIterator) ValidateData() error {
	if !it.ReflectValue.IsValid() {
		it.ReflectValue = reflect.ValueOf(it.Value)
	}

	if !it.ReflectValue.IsValid() {
		return ErrNoIterValue
	}

	if it.ReflectValue.Kind() != reflect.Ptr || it.ReflectValue.Elem().Kind() != reflect.Struct {
		return ErrInputIsNotPointerOrStruct
	}

	return nil
}

// ShouldSkipField returns true if field should be skipped.
//
// Field will be skipped if it is not valid / cannot be set, or if fieldName is equal to SkipFieldTagValue.
func ShouldSkipField(field reflect.Value, fieldName string, noUpdate bool) bool {
	return (!field.IsValid() || !field.CanSet()) || // Value is already set or not settable at all, skip it.
		IsTagSkip([]string{fieldName}) || // Field name is the one that should be skipped.
		(noUpdate && !field.IsZero()) // No update requested and field is already set.
}

// FieldNameWithSeparator will return FieldNameFunc that will use provided struct tag to get field name.
//
// Outer and inner field names will be concatenated with provided separator.
//
// Optionally result modifying functions can be specified with opts argument.
func FieldNameWithSeparator(tag, separator string, opts ...func(string) string) FieldNameFunc {
	return func(outerName string, currentField reflect.StructField) string {
		tagValue := TagValue(currentField, tag)[0]

		if strings.Contains(tagValue, ",") {
			tagValue = tagValue[:strings.Index(tagValue, ",")]
		}

		result := JoinFieldNames(outerName, tagValue, separator)
		for i := range opts {
			result = opts[i](result)
		}

		return result
	}
}

//nolint:golint
func PlainFieldNameWithPath(outer string, currentField reflect.StructField) string {
	if outer == "" {
		return currentField.Name
	}

	return outer + "." + currentField.Name
}

// JoinFieldNames will return concatenated string with 'separator' based on the value of 'outer'.
//
// If 'inner' == "-" - then "-" will be returned.
// If 'outer' == "" - then 'inner' will be returned as is.
// If 'inner' == "" - then outer + separator will be returned.
//
// Separator can be empty.
func JoinFieldNames(outer, inner, separator string) string {
	if inner == "-" {
		return "-"
	}

	if outer == "" {
		return inner
	}

	return outer + separator + inner
}

// TagValue returns splitted by comma value of tag for specified key.
//
// If no tags are provided - lowercased field name is returned.
//
// Note: If no such key in tag is present - DefaultConfigTag will be used to get value:
//
//	TagValue(field, "key") == TagValueByKeys(field, "key", DefaultConfigTag)
//
// See TagValueByKeys for usage examples.
func TagValue(field reflect.StructField, key string) []string {
	vals := TagValueByKeys(field.Tag, key, DefaultConfigTag)
	if vals == nil {
		return []string{strings.ToLower(field.Name)}
	}

	return vals
}

// TagValueByKeys returns comma-separated list of values based on keys and bool value representing if tag was found.
//
// Keys are sorted, which means that if first key has valid value - that value will be returned and
// no further keys will be checked.
//
// This function is useful to get default tag value if no requested tags have any values:
//
//	tagVal := TagValueByKeys(field, "highestPriority", "lowerPriority", "lowestPriority", "defaultTagKey")
//
// It is valid to call this function with zero or one key.
//
// If no tag found - nil is returned.
func TagValueByKeys(tag reflect.StructTag, keys ...string) []string {
	var tagValue string

	var ok bool

	for _, tagName := range keys {
		tagValue, ok = tag.Lookup(tagName)
		if ok {
			break
		}
	}

	if !ok {
		return nil
	}

	return strings.Split(tagValue, ",")
}

// IsTagSkip returns true if tag value was exactly "-".
func IsTagSkip(tagVals []string) bool {
	if len(tagVals) != 1 {
		return false
	}

	return tagVals[0] == SkipFieldTagValue
}

// IsTagOmitted returns true if tag was not found.
func IsTagOmitted(tagVals []string) bool {
	return len(tagVals) == 0
}

// IsStruct returns true if specified type is struct.
//
// Note: time.Time is not a struct by definition of this function!
func IsStruct(typ reflect.Type) bool {
	return typ.Kind() == reflect.Struct && typ != TimeType
}

//nolint:golint
func GetReflectElem(val interface{}) (reflect.Value, error) {
	refVal := reflect.ValueOf(val)

	if refVal.Kind() != reflect.Ptr {
		return reflect.Value{}, errors.New("value is not pointer")
	}

	return refVal.Elem(), nil
}

func GetEnvWithFallback(name string, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}

	return fallback
}
