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
		FieldNameFunc: e.FieldNameFunc,
		IteratorFunc:  e.IteratorFunc,
	}

	return it.Iterate()
}

// FieldNameFunc returns a field function which will get name from `env` tag,
// concatenated with '_'(underscore) and uppercased.
func (e Env) FieldNameFunc(outer string, field reflect.StructField) string {
	return internal.FieldNameWithSeparator(EnvTag, "_", strings.ToUpper)(outer, field)
}

// IteratorFunc sets a field to a value from environment.
//
// If field is not defined in environment - it is no-op.
func (e Env) IteratorFunc(fieldName string, field reflect.Value) error {
	val, ok := os.LookupEnv(fieldName)
	if !ok {
		return nil
	}

	return internal.SetReflectValueString(fieldName, val, field)
}
