package igconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// isTrue compares a string to determine the boolean value
func isTrue(substring string) bool {
	return strings.Contains(";TRUE;T;.T.;YES;Y;1;JA;J;", ";"+strings.ToUpper(substring)+";")
}

// setValue sets a value in the config struct
func (m *myConfig) setValue(v string) {
	k := m.f.Name
	val := reflect.ValueOf(m.c).Elem().FieldByName(k)

	switch m.f.Type.Kind() {
	case reflect.Bool:
		val.SetBool(isTrue(v))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			m.warnings = append(m.warnings, fmt.Sprintf("LoadConfig: value for field %s not a valid integer", k))
		} else {
			val.SetInt(n)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			m.warnings = append(m.warnings, fmt.Sprintf("LoadConfig: value for field %s not a valid unsigned integer", k))
		} else {
			val.SetUint(n)
		}

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			m.warnings = append(m.warnings, fmt.Sprintf("LoadConfig: value for field %s not a valid float", k))
		} else {
			val.SetFloat(n)
		}

	case reflect.String:
		val.SetString(v)

	default:
		m.warnings = append(m.warnings, fmt.Sprintf("LoadConfig: field %s unsupported type %s", m.f.Name, m.f.Type.Name()))
	}
}
