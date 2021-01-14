package internal

import (
	"errors"
	"reflect"
	"strings"
)

const DefaultConfigTag = "cfg"

const SkipField = "-"

var ErrInputIsNotPointerOrStruct = errors.New("input value is not struct or not pointer")
var ErrNoIterValue = errors.New("no value to iterate")

type FieldNameFunc func(outerName string, currentField reflect.StructField) string
type IteratorFunc func(fieldName string, field reflect.Value) error

type StructIterator struct {
	Value        interface{}
	ReflectValue reflect.Value
	BaseName     string
	// NoUpdate will specify that IteratorFunc cannot be run on fields with non-zero values.
	NoUpdate      bool
	FieldNameFunc FieldNameFunc
	IteratorFunc  IteratorFunc
}

func (it StructIterator) Iterate() error {
	if !it.ReflectValue.IsValid() {
		it.ReflectValue = reflect.ValueOf(it.Value)
	}

	if !it.ReflectValue.IsValid() {
		return ErrNoIterValue
	}

	if it.ReflectValue.Kind() != reflect.Ptr || it.ReflectValue.Elem().Kind() != reflect.Struct {
		return ErrInputIsNotPointerOrStruct
	}

	val := it.ReflectValue.Elem()
	valType := val.Type()

	for i := 0; i < valType.NumField(); i++ {
		toField := val.Field(i)
		if !toField.IsValid() || !toField.CanSet() {
			// Value is already set or not settable at all, skip it.
			continue
		}

		fieldName := it.FieldNameFunc(it.BaseName, valType.Field(i))
		if fieldName == SkipField { // This is the sign that this field should be skipped.
			continue
		}

		// Set zero-value to pointer fields when they are being processed.
		if toField.Kind() == reflect.Ptr && toField.IsNil() {
			toField.Set(reflect.New(toField.Type().Elem()))

			toField = toField.Elem()
		}

		// If it is a struct - try to set it's inner fields to.
		if toField.Kind() == reflect.Struct && toField.Type() != TimeType {
			subIter := StructIterator{
				ReflectValue:  toField.Addr(), // This is just simple solution for using pointer structs as inputs.
				BaseName:      fieldName,
				FieldNameFunc: it.FieldNameFunc,
				IteratorFunc:  it.IteratorFunc,
			}

			if err := subIter.Iterate(); err != nil {
				return err
			}

			continue
		}

		if it.NoUpdate && !toField.IsZero() { // Do not update non-empty fields if requested.
			continue
		}

		if err := it.IteratorFunc(fieldName, toField); err != nil {
			return err
		}
	}

	return nil
}

// FieldNameWithModifiers will run fieldNameGetter to receive field name and
// then will apply all funcs m to resulting field name.
func FieldNameWithModifiers(fieldNameGetter FieldNameFunc, m ...func(string) string) FieldNameFunc {
	return func(outer string, inner reflect.StructField) string {
		fieldName := fieldNameGetter(outer, inner)

		for i := range m {
			fieldName = m[i](fieldName)
		}

		return fieldName
	}
}

func CombineFieldNameFuncs(fncs ...FieldNameFunc) FieldNameFunc {
	return func(outerName string, currentField reflect.StructField) string {
		for i := range fncs {
			if fncs[i] == nil {
				continue
			}

			fieldName := fncs[i](outerName, currentField)
			if fieldName != "" {
				return fieldName
			}
		}

		return ""
	}
}

func EnvFieldName(outer string, currentField reflect.StructField) string {
	tagValues := TagValue(currentField, "env")
	if tagValues == nil {
		return "-"
	}

	tagValue := tagValues[0]
	if strings.Contains(tagValue, ",") {
		tagValue = tagValue[:strings.Index(tagValue, ",")]
	}

	if outer == "" {
		return strings.ToUpper(tagValue)
	}

	return strings.ToUpper(outer + "_" + tagValue)
}

func PlainFieldName(_ string, currentField reflect.StructField) string {
	return currentField.Name
}

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
//	TagValue(field, "key") == TagValueByKeys(field, "key", DefaultConfigTag)
//
// See TagValueByKeys for usage examples.
func TagValue(field reflect.StructField, key string) []string {
	vals := TagValueByKeys(field, key, DefaultConfigTag)
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
//	tagVal := TagValueByKeys(field, "highestPriority", "lowerPriority", "lowestPriority", "defaultTagKey")
//
// It is valid to call this function with zero or one key.
//
// If no tag found - nil is returned.
func TagValueByKeys(field reflect.StructField, keys ...string) []string {
	var tagValue string
	var ok bool
	for _, tagName := range keys {
		tagValue, ok = field.Tag.Lookup(tagName)
		if ok {
			break
		}

	}

	if !ok {
		return nil
	}

	return strings.Split(tagValue, ",")
}

// IsTagSkip returns true if tag value was exactly "-"
func IsTagSkip(tagVals []string) bool {
	if len(tagVals) != 1 {
		return false
	}

	return tagVals[0] == SkipField
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

func GetReflectElem(val interface{}) (reflect.Value, error) {
	refVal := reflect.ValueOf(val)

	if refVal.Kind() != reflect.Ptr {
		return reflect.Value{}, errors.New("value is not pointer")
	}

	return refVal.Elem(), nil
}
