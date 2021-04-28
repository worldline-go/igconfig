package loader_test

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

func TestReader_LoadWorkDir_EnvFile(t *testing.T) {
	data := `salary: 112.34
host: example.com
innerstruct:
  str: test_me`

	f, err := os.CreateTemp("", "conf")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(data)
	require.NoError(t, err)

	os.Setenv(loader.EnvConfigFile, f.Name())
	defer os.Unsetenv(loader.EnvConfigFile)

	r := loader.Reader{}

	var c testdata.TestConfig
	require.NoError(t, r.LoadWorkDir("a", &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      112.34,
		Host:        "example.com",
		InnerStruct: testdata.InnerStruct{Str: "test_me"},
	}, c)
}

func TestReader_LoadWorkDir(t *testing.T) {
	data := `salary: 112.34
host: example.com
innerstruct:
  str: test_me`

	wd, _ := os.Getwd()
	f, err := os.Create(path.Join(wd, "app"+loader.ConfFileSuffix))
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(data)
	require.NoError(t, err)

	r := loader.Reader{}

	var c testdata.TestConfig
	require.NoError(t, r.LoadWorkDir("app", &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      112.34,
		Host:        "example.com",
		InnerStruct: testdata.InnerStruct{Str: "test_me"},
	}, c)
}

func TestReader_LoadWorkDir_DoesNotExist(t *testing.T) {
	r := loader.Reader{}

	err := r.LoadWorkDir("app", &testdata.TestConfig{})
	assert.True(t, errors.Is(err, os.ErrNotExist))
}
