package loader

import (
	"fmt"
	"reflect"
	"strconv"
)

// setValue sets a value in the config struct
func setValue(fieldName, v string, to reflect.Value) error {
	const (
		funcName = "setValue"
		valueMsg = "%s: value for field %s not a valid %s: %q"
	)

	kindName := to.Type().String()

	switch to.Kind() {
	case reflect.Bool:
		res, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf(valueMsg, funcName, fieldName, kindName, v)
		}

		to.SetBool(res)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, funcName, fieldName, kindName, v)
		}
		to.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, funcName, fieldName, kindName, v)
		}
		to.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, funcName, fieldName, kindName, v)
		}
		to.SetFloat(n)
	case reflect.String:
		to.SetString(v)
	default:
		return fmt.Errorf("%s: field %s unsupported type %s", funcName, fieldName, kindName)
	}

	return nil
}
