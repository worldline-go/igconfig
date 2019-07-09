package igconfig

import (
	"flag"
	"os"
	"reflect"
	"strings"
	"strconv"
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
	"gitlab.test.igdcs.com/finops/utils/basics/igstrings.git"
)

// quietFlagSet overrides the standard FlagSet to silence the Usage function
type quietFlagSet struct {
	flag.FlagSet
}

// Usage overrides the corresponding function in FlagSet to silence it
func (f *quietFlagSet) Usage() {
	// do nothing
}

// loadCmdline loads config values from the command line
func loadCmdline (c interface{}) {
	iglog.DebugFull("LoadConfig: loading command-line parameters")

	var flags quietFlagSet

	e := reflect.ValueOf(c).Elem()
	t := e.Type()

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

	flags.Parse(os.Args[1:])

	for i := 0; i < t.NumField(); i++ {
		f   := t.Field(i)
		val := e.FieldByName(f.Name)

		var nn     []string
		var newVal string

		if tag, ok := f.Tag.Lookup("cmd"); ok {
			nn = strings.Split(tag, ",")
		}

		if n := strings.ToLower(f.Name); !igstrings.SliceContains(nn, n) {
			nn = append(nn, n)
		}

		for _, n := range nn {
			flg := flags.Lookup(n)
			
			if flg != nil {
			    v := flg.Value.String()
				switch f.Type.Kind() {
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
			setValue(c, f, newVal)
		}
	}
}
