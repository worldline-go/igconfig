package loader

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/internal"
)

var _ Loader = Flags{}

const CmdTag = "cmd"

// Flags will parse CMD args.
//
// Breaking change from v1: parsing will fail with usage if unknown flag will be found.
type Flags struct {
	// NoUsage may be used to silence usage on invalid flags.
	NoUsage bool
}

func (f Flags) Load(_ string, to interface{}) error {
	return f.LoadSlice(to, os.Args[1:])
}

// LoadCmdline loads config values from the command line.
func (f Flags) LoadSlice(to interface{}, args []string) error {
	if len(args) == 0 {
		return nil
	}

	flags := flag.FlagSet{}

	if f.NoUsage {
		flags.Usage = func() {}
	}

	// This is the function to add flags.
	addFlags := func(fieldName string, field reflect.Value) error {
		setFlagForKind(&flags, field.Type().Kind(), fieldName, field)
		return nil
	}
	// This is the function to set flag values
	processFlags := func(fieldName string, field reflect.Value) error {
		// Flag will always be defined, as guaranteed by previous iteration.
		fl := flags.Lookup(fieldName)
		if _, ok := fl.Value.(flag.Getter).Get().(reflect.Value); ok {
			// If value is reflect.Value then it should not be set.
			// It is already set when the flags were parsed
			return nil
		}

		return internal.SetReflectValueString(fieldName, fl.Value.String(), field)
	}

	it := internal.StructIterator{
		Value: to,
		FieldNameFunc: func(outerName string, currentField reflect.StructField) string {
			tags := internal.TagValue(currentField, CmdTag)
			if tags == nil {
				return internal.SkipFieldTagValue
			}

			return strings.ToLower(internal.JoinFieldNames(outerName, tags[0], "-"))
		},
		IteratorFunc: addFlags,
	}
	if err := it.Iterate(); err != nil {
		return err
	}

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("flags parsing error: %s", err.Error())
	}

	it.IteratorFunc = processFlags
	return it.Iterate()
}

func setFlagForKind(flags *flag.FlagSet, fieldKind reflect.Kind, flagName string, defValue reflect.Value) {
	if setter := internal.GetCustomSetter(defValue.Type()); setter != nil {
		flags.Var(CustomVar{Setter: setter, Val: defValue}, flagName, "")

		return
	}

	switch fieldKind {
	case reflect.Bool:
		flags.Bool(flagName, defValue.Bool(), "")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flags.Int64(flagName, defValue.Int(), "")
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		flags.Uint64(flagName, defValue.Uint(), "")
	case reflect.Float32, reflect.Float64:
		flags.Float64(flagName, defValue.Float(), "")
	case reflect.String:
		flags.String(flagName, defValue.String(), "")
	}
}

// errorFromSlice will return single error by concatenating all of the provided errors.
//
// If no errors are provided - nil is returned.
func errorFromSlice(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var errString string

	for i, err := range errs {
		if i != 0 {
			errString += ", "
		}

		errString += err.Error()
	}

	return errors.New(errString)
}

type CustomVar struct {
	Setter internal.TypeSetter
	Val    reflect.Value
}

func (c CustomVar) String() string {
	if !c.Val.IsValid() || c.Val.IsZero() {
		return ""
	}

	return fmt.Sprint(c.Val.Interface())
}

func (c CustomVar) Set(s string) error {
	if s == "" {
		return nil
	}

	return c.Setter(s, c.Val)
}

// Get is necessary to get reflect.Value as is and not as a string.
func (c CustomVar) Get() interface{} {
	return c.Val
}
