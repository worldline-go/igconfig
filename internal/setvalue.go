package internal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	AddCustomType(time.Time{}, func(input string, val reflect.Value) error {
		t, err := time.ParseInLocation(time.RFC3339, input, time.Local)
		if err != nil {
			return err
		}

		val.Set(reflect.ValueOf(t))

		return nil
	})

	AddCustomType(time.Duration(0), func(input string, val reflect.Value) error {
		d, err := time.ParseDuration(input)
		if err != nil {
			return err
		}

		val.Set(reflect.ValueOf(d))

		return nil
	})

	AddCustomType([]string(nil), func(input string, val reflect.Value) error {
		if input == "" {
			return nil
		}

		if !strings.Contains(input, SliceSeparator) {
			val.Set(reflect.ValueOf([]string{input}))
		}

		sl := strings.Split(input, SliceSeparator)
		val.Set(reflect.ValueOf(sl))

		return nil
	})
}

type TypeSetter func(input string, val reflect.Value) error

var SliceSeparator = ","

var typeMap = map[reflect.Type]TypeSetter{}
var typeMapMu sync.RWMutex

var TimeType = reflect.TypeOf(time.Time{})

// isTrue compares a string to determine the boolean value.
//
// Returns true only if string can be parsed as true.
// Otherwise returns false(even if value can't be parsed boolean).
func isTrue(str string) bool {
	res, _ := strconv.ParseBool(str)

	return res
}

// AddCustomType will add custom setter for specified type.
// It is rare to add setter for pointer type: only to satisfy 'flags' loader.
// Warning: pointer types are not guaranteed to work!
//
// This types are also used for flag setters.
//
// To disable custom setter - call this method with setter == nil.
// This is also the method to remove pointer type setter provided by this function.
//
// For example:
//	// This will add custom setter for type 'int64'
//	AddCustomType(int64(0), someSetter)
//
//	// To disable any custom setter - provide nil as setter for required type:
//	AddCustomType(int64(0), nil)
//
//	// To add pointer-only type use this:
//	AddCustomType((*time.Duration)(nil), someDurationSetter)
//	// This will add custom setter for type '*time.Duration' only.
func AddCustomType(typ interface{}, setter TypeSetter) {
	typeMapMu.Lock()

	reflTyp := reflect.TypeOf(typ)

	typeMap[reflTyp] = setter

	typeMapMu.Unlock()
}

func GetCustomSetter(typ reflect.Type) TypeSetter {
	typeMapMu.RLock()
	defer typeMapMu.RUnlock()

	return typeMap[typ]
}

func SetStructFieldValue(fieldName, v string, strct reflect.Value) error {
	strct = reflect.Indirect(strct)

	field := strct.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("set: non-existing field passed: %s", fieldName)
	}

	return SetReflectValueString(fieldName, v, field)
}

// SetReflectValueString sets a value in the config struct.
//
// 'val' is actual variable that should be set.
func SetReflectValueString(fieldName, v string, val reflect.Value) error {
	const valueMsg = "value for val %q not a valid %q"

	kindName := val.Type().String()

	if setter := GetCustomSetter(val.Type()); setter != nil {
		return setter(v, val)
	}

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
