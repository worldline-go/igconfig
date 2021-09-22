package internal

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TypeSetter func(input string, val reflect.Value) error

var SliceSeparator = ","

var typeMap = map[reflect.Type]TypeSetter{
	reflect.TypeOf(time.Time{}): func(input string, val reflect.Value) error {
		t, err := time.ParseInLocation(time.RFC3339, input, time.Local)
		if err != nil {
			return err
		}

		val.Set(reflect.ValueOf(t))

		return nil
	},
	reflect.TypeOf(time.Duration(0)): func(input string, val reflect.Value) error {
		d, err := time.ParseDuration(input)
		if err != nil {
			return err
		}

		val.Set(reflect.ValueOf(d))

		return nil
	},
	reflect.TypeOf([]string(nil)): func(input string, val reflect.Value) error {
		if input == "" {
			return nil
		}

		if !strings.Contains(input, SliceSeparator) {
			val.Set(reflect.ValueOf([]string{input}))
		}

		sl := strings.Split(input, SliceSeparator)
		val.Set(reflect.ValueOf(sl))

		return nil
	},
}
var typeMapMu sync.RWMutex

var TimeType = reflect.TypeOf(time.Time{})

var unmarshalTextType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// isTrue compares a string to determine the boolean value.
//
// Returns true only if string can be parsed as true.
// Otherwise returns false(even if value can't be parsed boolean).
func isTrue(str string) bool {
	res, _ := strconv.ParseBool(str)

	return res
}

// AddCustomTypeSetter will add custom setter for specified type.
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
//	AddCustomTypeSetter(int64(0), someSetter)
//
//	// To disable any custom setter - provide nil as setter for required type:
//	AddCustomTypeSetter(int64(0), nil)
//
//	// To add pointer-only type use this:
//	AddCustomTypeSetter((*time.Duration)(nil), someDurationSetter)
//	// This will add custom setter for type '*time.Duration' only.
func AddCustomTypeSetter(typ interface{}, setter TypeSetter) {
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
//
// Custom types can be added for value setting with AddCustomTypeSetter.
//
// If type that is being set implements encoding.TextUnmarshaler - it will be used instead of direct field set.
// If custom setter function is present AND type implements encoding.TextUnmarshaler -
// custom setter function will take priority.
func SetReflectValueString(fieldName, v string, val reflect.Value) error {
	const valueMsg = "value for val %q not a valid %q"

	val = reflect.Indirect(val)

	kindName := val.Type().String()

	if setter := GetCustomSetter(val.Type()); setter != nil {
		return setter(v, val)
	}
	// Check if this type or pointer to this type implements encoding.TextUnmarshaler
	if implVal := TypeImplementsInterface(val, unmarshalTextType); implVal.IsValid() {
		return implVal.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(v))
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

// TypeImplementsInterface checks if provided value's type,
// or pointer(if possible to make pointer) to it's type implements specified interface type.
//
// Argument interfaceType can be created as reflect.TypeOf((*interface{})(nil)).Elem().
func TypeImplementsInterface(val reflect.Value, interfaceType reflect.Type) reflect.Value {
	if val.Type().Implements(interfaceType) {
		return val
	}

	if val.CanAddr() && val.Addr().Type().Implements(interfaceType) {
		return val.Addr()
	}

	return reflect.Value{}
}
