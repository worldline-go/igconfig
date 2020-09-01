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

func (e Env) Load(_ string, to interface{}) error {
	refVal, err := internal.GetReflectElem(to)
	if err != nil {
		return err
	}

	return e.ReflectLoad("", refVal)
}

func (e Env) ReflectLoad(_ string, to reflect.Value) error {
	t := to.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tags := internal.TagValue(field, EnvTag)

		for _, name := range tags {
			val, ok := lookupEnv(name)
			if !ok {
				continue
			}

			if err := internal.SetReflectValue(field.Name, val, to.Field(i)); err != nil {
				return err
			}

			break
		}
	}

	return nil
}

// lookupEnv tests for an environment variable(only upper case), and if found - returns its value.
//
// Second return value tells if such environmental variable exists.
func lookupEnv(envName string) (string, bool) {
	return os.LookupEnv(strings.ToUpper(envName))
}
