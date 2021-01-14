package loader_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

func TestCmdlineValues(t *testing.T) {
	args := []string{"-name", "Piet", "--port", "1234", "--host=bol.com",
		"--age", "25", "--salary", "1500.00", "--dur", "94s", "--secure",
		"--innerstruct-dur", "14s",
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
		Unused: nil,
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
