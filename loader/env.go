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
		FieldNameFunc: internal.FieldNameWithSeparator(EnvTag, "_", strings.ToLower),
		IteratorFunc: func(fieldName string, field reflect.Value) error {
			val, ok := os.LookupEnv(strings.ToUpper(fieldName))
			if !ok {
				return nil
			}

			return internal.SetReflectValueString(fieldName, val, field)
		},
	}

	return it.Iterate()
}
