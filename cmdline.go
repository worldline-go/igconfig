package igconfig

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// loadCmdline loads config values from the command line
func (m *localData) loadCmdline(args []string) error {
	if len(args) == 0 {
		return nil
	}

	t := m.userStruct.Type()

	flags := flag.FlagSet{Usage: func() {}}
	argToFieldName := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var nn []string

		if tag, ok := field.Tag.Lookup("cmd"); ok {
			nn = strings.Split(tag, ",")
		}

		for _, n := range nn {
			val := m.userStruct.FieldByName(field.Name)

			switch field.Type.Kind() {
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
			argToFieldName[n] = field.Name
		}
	}

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("loadCmdline error parsing parameters: %s", err.Error())
	}

	flags.Visit(func(fl *flag.Flag) {
		m.setValue(argToFieldName[fl.Name], fl.Value.String())
	})

	return nil
}
