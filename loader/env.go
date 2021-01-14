package loader

import (
	"os"
	"reflect"
	"strings"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/internal"
)

var _ Loader = Env{}

const EnvTag = "env"

// Env allows to load values from environmental variables.
//
// For consistency and best-practices of Linux environmental variables
// only upper-case values will be used. Even if in tag it is specified in lower-case.
//
// Breaking change from v1: variable name will be upper-cased when doing lookup. No other cases are checked.
type Env struct{}

func (e Env) Load(baseName string, to interface{}) error {
	it := internal.StructIterator{
		Value:         to,
		BaseName:      baseName,
		FieldNameFunc: internal.EnvFieldName,
		IteratorFunc: func(fieldName string, field reflect.Value) error {
			val, ok := lookupEnv(fieldName)
			if !ok {
				return nil
			}

			if err := internal.SetReflectValueString(fieldName, val, field); err != nil {
				return err
			}

			return nil
		},
	}

	return it.Iterate()
}

// lookupEnv tests for an environment variable(only upper case), and if found - returns its value.
//
// Second return value tells if such environmental variable exists.
func lookupEnv(envName string) (string, bool) {
	return os.LookupEnv(strings.ToUpper(envName))
}
