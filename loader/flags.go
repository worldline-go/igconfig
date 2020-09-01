package loader

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"

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

	refVal, err := internal.GetReflectElem(to)
	if err != nil {
		return err
	}

	t := refVal.Type()

	flags := flag.FlagSet{}

	if f.NoUsage {
		flags.Usage = func() {}
	}

	argToFieldName := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tags := internal.TagValue(field, CmdTag)

		for _, flagName := range tags {
			val := refVal.FieldByName(field.Name)

			setFlagForKind(&flags, field.Type.Kind(), flagName, val)

			argToFieldName[flagName] = field.Name
		}
	}

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("flags parsing error: %s", err.Error())
	}

	var errs []error

	flags.Visit(func(fl *flag.Flag) {
		if err := internal.SetStructFieldValue(argToFieldName[fl.Name], fl.Value.String(), refVal); err != nil {
			errs = append(errs, err)
		}
	})

	return errorFromSlice(errs)
}

func setFlagForKind(flags *flag.FlagSet, fieldKind reflect.Kind, flagName string, defValue reflect.Value) {
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
