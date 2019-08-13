package igconfig

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gitlab.test.igdcs.com/finops/utils/basics/igstrings.git"
)

// loadCmdline loads config values from the command line
func (m *localData) loadCmdline(args []string) error {
	if len(args) == 0 {
		return nil
	}

	v := reflect.ValueOf(m.userStruct)
	e := v.Elem()
	t := e.Type()

	flags := flag.FlagSet{Usage: func() {}}

	for i := 0; i < t.NumField(); i++ {
		m.fld = t.Field(i)
		var nn []string

		if tag, ok := m.fld.Tag.Lookup("cmd"); ok {
			nn = strings.Split(tag, ",")
		}

		for _, n := range nn {
			val := e.FieldByName(m.fld.Name)

			switch m.fld.Type.Kind() {
			case reflect.Bool:
				flags.Bool(n, val.Bool(), "")
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				flags.Int64(n, val.Int(), "")
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				flags.Uint64(n, val.Uint(), "")
			case reflect.Float32, reflect.Float64:
				flags.Float64(n, val.Float(), "")
			case reflect.String:
				flags.String(n, val.String(), "")
			}
		}
	}

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("loadCmdline error parsing parameters: %s", err.Error())
	}

	for i := 0; i < t.NumField(); i++ {
		m.fld = t.Field(i)
		val := e.FieldByName(m.fld.Name)

		var nn []string
		var newVal string

		if tag, ok := m.fld.Tag.Lookup("cmd"); ok {
			nn = strings.Split(tag, ",")
		}

		if n := strings.ToLower(m.fld.Name); !igstrings.SliceContains(nn, n) {
			nn = append(nn, n)
		}

		for _, n := range nn {
			flg := flags.Lookup(n)

			if flg != nil {
				v := flg.Value.String()
				switch m.fld.Type.Kind() {
				case reflect.Bool:
					b := isTrue(v)
					if b != val.Bool() {
						newVal = v
					}

				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					i, err := strconv.ParseInt(v, 0, 64)
					if err == nil && i != val.Int() {
						newVal = v
					}

				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					i, err := strconv.ParseUint(v, 0, 64)
					if err == nil && i != val.Uint() {
						newVal = v
					}

				case reflect.Float32, reflect.Float64:
					i, err := strconv.ParseFloat(v, 64)
					if err == nil && i != val.Float() {
						newVal = v
					}

				case reflect.String:
					if v != val.String() {
						newVal = v
					}
				}
			}
		}

		if newVal != "" {
			m.setValue(newVal)
		}
	}

	return nil
}
