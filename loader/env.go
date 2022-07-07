package loader

import (
	"context"
	"os"
	"reflect"
	"strings"

	"github.com/worldline-go/igconfig/internal"
)

var _ Loader = Env{}

// EnvTag is a tag name for environment variable.
const EnvTag = "env"

// Env allows to load values from environmental variables.
//
// For consistency and best-practices of Linux environmental variables
// only upper-case values will be used. Even if in tag it is specified in lower-case.
//
// Variable name will be upper-cased when doing lookup. No other cases are checked.
type Env struct{}

// LoadWithContext implementation for Env ignores variable name prefix.
// We want to load env vars without Vault like prefix.
func (l Env) LoadWithContext(_ context.Context, _ string, to interface{}) error {
	it := internal.StructIterator{
		Value:         to,
		FieldNameFunc: l.FieldNameFunc,
		IteratorFunc:  l.IteratorFunc,
	}

	return it.Iterate()
}

// Load is just same as LoadWithContext without context.
func (l Env) Load(_ string, to interface{}) error {
	return l.LoadWithContext(context.TODO(), "", to)
}

// FieldNameFunc returns a field function which will get name from `env` tag,
// concatenated with '_'(underscore) and uppercased.
func (l Env) FieldNameFunc(outer string, field reflect.StructField) string {
	return internal.FieldNameWithSeparator(EnvTag, "_", strings.ToUpper)(outer, field)
}

// IteratorFunc sets a field to a value from environment.
//
// If field is not defined in environment - it is no-op.
func (l Env) IteratorFunc(fieldName string, field reflect.Value) error {
	val, ok := os.LookupEnv(fieldName)
	if !ok {
		return nil
	}

	return internal.SetReflectValueString(fieldName, val, field)
}
