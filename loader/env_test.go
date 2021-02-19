package loader_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

func TestEnvValues(t *testing.T) {
	const funcName = "TestEnvValues"

	if err := os.Setenv("HOST", "127.0.0.1"); err != nil {
		t.Errorf("%s could not set environment variable 'HOST'", funcName)
	}
	if err := os.Setenv("PORT", "12345"); err != nil {
		t.Errorf("%s could not set environment variable 'Port'", funcName)
	}
	if err := os.Setenv("age", "44"); err != nil {
		t.Errorf("%s could not set environment variable 'age'", funcName)
	}
	if err := os.Setenv("address", "should_not_be_set"); err != nil {
		t.Errorf("%s could not set environment variable 'address'", funcName)
	}
	if err := os.Setenv("INNERSTRUCT_STRING", "hello"); err != nil {
		t.Errorf("%s could not set environment variable 'INNERSTRUCT_STRING'", funcName)
	}
	if err := os.Setenv("SLICE", "3,4,4"); err != nil {
		t.Errorf("%s could not set environment variable 'SLICE'", funcName)
	}

	var c testdata.TestConfig

	require.NoError(t, (loader.Default{}).Load("", &c))

	assert.NoError(t, (loader.Env{}).Load("", &c))

	assert.Equal(t, testdata.TestConfig{
		Name:    "Jan",
		Age:     18,
		Address: "localhost",
		Salary:  2000.0,
		Host:    "127.0.0.1",
		Port:    12345,
		Secure:  false,
		Slice:   []string{"3", "4", "4"},
		InnerStruct: testdata.InnerStruct{
			Str:  "hello",
			Time: testdata.ParsedTime,
		},
	}, c)
}
