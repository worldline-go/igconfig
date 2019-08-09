package igconfig

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gitlab.test.igdcs.com/finops/utils/basics/igstrings.git"
)

// loadCmdline loads config values from the command line
func (m *myConfig) loadCmdline() {
	const funcName = "loadCmdline"

	v := reflect.ValueOf(m.c)
	if v.Kind() != reflect.Ptr {
		return
	}

	e := v.Elem()
	t := e.Type()
	if t.Kind() != reflect.Struct {
		return
	}

	flags := flag.FlagSet{Usage: func() {}}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		var nn []string

		if tag, ok := f.Tag.Lookup("cmd"); ok {
			nn = strings.Split(tag, ",")
		}

		for _, n := range nn {
			val := e.FieldByName(f.Name)

			switch f.Type.Kind() {
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

	if err := flags.Parse(os.Args[1:]); err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("%s could not parse command-line parameters: %s", funcName, err.Error()))
	}

	for i := 0; i < t.NumField(); i++ {
		m.f = t.Field(i)
		val := e.FieldByName(m.f.Name)

		var nn []string
		var newVal string

		if tag, ok := m.f.Tag.Lookup("cmd"); ok {
			nn = strings.Split(tag, ",")
		}

		if n := strings.ToLower(m.f.Name); !igstrings.SliceContains(nn, n) {
			nn = append(nn, n)
		}

		for _, n := range nn {
			flg := flags.Lookup(n)

			if flg != nil {
				v := flg.Value.String()
				switch m.f.Type.Kind() {
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
}
