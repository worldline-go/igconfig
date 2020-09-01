package internal

import (
	"fmt"
	"reflect"
	"strconv"
)

// isTrue compares a string to determine the boolean value.
//
// Returns true only if string can be parsed as true.
// Otherwise returns false(even if value can't be parsed boolean).
func isTrue(str string) bool {
	res, _ := strconv.ParseBool(str)

	return res
}

func SetStructFieldValue(fieldName, v string, strct reflect.Value) error {
	strct = reflect.Indirect(strct)

	field := strct.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("set: non-existing field passed: %s", fieldName)
	}

	return SetReflectValue(fieldName, v, field)
}

// SetReflectValue sets a value in the config struct.
//
// 'val' is actual variable that should be set.
func SetReflectValue(fieldName, v string, val reflect.Value) error {
	const valueMsg = "value for val %q not a valid %q"

	kindName := val.Type().String()

	switch val.Kind() {
	case reflect.Bool:
		val.SetBool(isTrue(v))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, fieldName, kindName)
		}

		val.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, fieldName, kindName)
		}

		val.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, fieldName, kindName)
		}

		val.SetFloat(n)
	case reflect.String:
		val.SetString(v)
	default:
		return fmt.Errorf("val %q has unsupported type %q", fieldName, kindName)
	}

	return nil
}
