package igconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// isTrue compares a string to determine the boolean value
func isTrue(substring string) bool {
	res, _ := strconv.ParseBool(substring)

	// if parsed to true - then no need to go further
	if res {
		return true
	}

	return res || strings.Contains(";T;.T.;YES;Y;JA;J;", ";"+strings.ToUpper(substring)+";")
}

// setValue sets a value in the config struct
func (m *localData) setValue(fieldName, v string) {
	const (
		funcName = "setValue"
		valueMsg = "%s: value for field %s not a valid %s"
	)

	val := m.userStruct.FieldByName(fieldName)
	if !val.IsValid() {
		m.messages = append(m.messages, fmt.Sprintf("%s: non-existing field passed: %s", funcName, fieldName))
		return
	}

	if err := setValue(fieldName, v, val); err != nil {
		m.messages = append(m.messages, err.Error())
	}
}

// setValue sets a value in the config struct
func setValue(fieldName, v string, to reflect.Value) error {
	const (
		funcName = "setValue"
		valueMsg = "%s: value for field %s not a valid %s"
	)

	kindName := to.Type().String()

	switch to.Kind() {
	case reflect.Bool:
		to.SetBool(isTrue(v))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, funcName, fieldName, kindName)
		}
		to.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, funcName, fieldName, kindName)
		}
		to.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf(valueMsg, funcName, fieldName, kindName)
		}
		to.SetFloat(n)
	case reflect.String:
		to.SetString(v)
	default:
		return fmt.Errorf("%s: field %s unsupported type %s", funcName, fieldName, kindName)
	}

	return nil
}
