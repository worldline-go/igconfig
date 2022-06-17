package loader_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/worldline-go/igconfig/internal"
	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/igconfig/testdata"
)

func TestCmdlineValues(t *testing.T) {
	args := []string{"-name", "Piet", "--port", "1234", "--host=bol.com",
		"--age", "25", "--salary", "1500.00", "--dur", "94s", "--secure",
		"--slice", "4,4,5", "--innerstruct-dur", "14s",
	}

	var c testdata.TestConfig

	assert.NoError(t, (loader.Flags{}).LoadSlice(&c, args))

	assert.Equal(t, testdata.TestConfig{
		Name:   "Piet",
		Age:    25,
		Salary: 1500.0,
		Host:   "bol.com",
		Port:   1234,
		Secure: true,
		Slice:  []string{"4", "4", "5"},
		Dur:    94 * time.Second,
		InnerStruct: testdata.InnerStruct{
			Dur: 14 * time.Second,
		},
	}, c)
}

func TestFlags_LoadSliceInvalid(t *testing.T) {
	assert.EqualError(t, (loader.Flags{NoUsage: true}).LoadSlice(&testdata.TestConfig{}, []string{"-x"}),
		"flags parsing error: flag provided but not defined: -x")
}

func TestFlags_FieldNameFunc(t *testing.T) {
	var s struct {
		One   bool     `cfg:"ones"`
		Two   string   `cmd:"2"`
		Three []string `cfg:"three" cmd:"-"`
		Inner struct {
			Field      bool `cfg:"f"`
			unexported bool `cfg:"unexported"`
		} `cmd:"in"`
	}

	var res []string
	it := internal.StructIterator{
		Value:         &s,
		FieldNameFunc: fieldCapture(loader.Flags{}.FieldNameFunc, &res),
		IteratorFunc:  func(fieldName string, field reflect.Value) error { return nil },
	}

	require.NoError(t, it.Iterate())

	require.Equal(t, []string{"ones", "2", "-", "in", "in-f"}, res)
}

func fieldCapture(nameFunc internal.FieldNameFunc, namesSlice *[]string) internal.FieldNameFunc {
	return func(outerName string, currentField reflect.StructField) string {
		name := nameFunc(outerName, currentField)

		*namesSlice = append(*namesSlice, name)

		return name
	}
}
