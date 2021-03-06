package loader_test

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/igconfig/testdata"
)

func TestFile_Load_EnvFile(t *testing.T) {
	data := `name: Holland
salary: 112.34
host: example.com
innerstruct:
  string: test_me`

	f, err := os.CreateTemp("", "*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name()) //nolint:errcheck

	_, err = f.WriteString(data)
	require.NoError(t, err)

	os.Setenv(loader.EnvConfigFile, f.Name()) //nolint:errcheck
	defer os.Unsetenv(loader.EnvConfigFile)   //nolint:errcheck

	l := loader.File{}

	var c testdata.TestConfig
	require.NoError(t, l.Load("a", &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      112.34,
		Host:        "example.com",
		InnerStruct: testdata.InnerStruct{Str: "test_me"},
	}, c)

	dataUgly := `salary 32131
host: dontuseme.com
innerstruct
  string: test_me`

	_, err = f.WriteString(dataUgly)
	require.NoError(t, err)

	require.Error(t, l.Load("a", &c))
}

func TestFile_LoadWorkDir(t *testing.T) {
	data := `name: Holland
salary: 112.34
host: example.com
innerstruct:
  string: test_me`

	wd, _ := os.Getwd()
	f, err := os.Create(path.Join(wd, "app.yaml"))
	require.NoError(t, err)
	defer os.Remove(f.Name()) //nolint:errcheck

	_, err = f.WriteString(data)
	require.NoError(t, err)

	l := loader.File{}

	var c testdata.TestConfig
	require.NoError(t, l.LoadWorkDir(" /pathUpper/test/app/", &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      112.34,
		Host:        "example.com",
		InnerStruct: testdata.InnerStruct{Str: "test_me"},
	}, c)
}

func TestFile_LoadWorkDir_DoesNotExist(t *testing.T) {
	l := loader.File{}

	err := l.LoadWorkDir("app", &testdata.TestConfig{})
	assert.True(t, errors.Is(err, loader.ErrNoConfFile))
}

func TestFile_CHECK_ETC(t *testing.T) {
	dataEtc := `name: Holland
salary: 999.9
host: exampleetc.com
innerstruct:
  string: test_me_ETC`

	fEtc, err := os.CreateTemp("", "*.yaml")
	require.NoError(t, err)
	defer os.Remove(fEtc.Name()) //nolint:errcheck

	_, err = fEtc.WriteString(dataEtc)
	require.NoError(t, err)

	l := loader.File{EtcPath: filepath.Dir(fEtc.Name())}

	var c testdata.TestConfig
	appnameWithSuffix := path.Base(fEtc.Name())
	appname := appnameWithSuffix[:strings.LastIndex(appnameWithSuffix, ".")]

	require.NoError(t, l.Load(appname, &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      999.9,
		Host:        "exampleetc.com",
		InnerStruct: testdata.InnerStruct{Str: "test_me_ETC"},
	}, c)

	c = testdata.TestConfig{}

	l.NoFolderCheck = true

	require.NoError(t, l.Load(appname, &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      0,
		Host:        "",
		InnerStruct: testdata.InnerStruct{Str: ""},
	}, c)
}

func TestFile_CHECK_ORDER(t *testing.T) {
	dataEtc := `{
"name": "Holland",
"salary": 999.9,
"host": "exampleetc.com",
"innerstruct": {
    "string": "test_me_ETC"
  }
}`

	fEtc, err := os.CreateTemp("", "*.json")
	require.NoError(t, err)
	defer os.Remove(fEtc.Name()) //nolint:errcheck

	_, err = fEtc.WriteString(dataEtc)
	require.NoError(t, err)

	data := `salary: 112.34
host: example.com
innerstruct:
  string: test_me`

	wd, _ := os.Getwd()
	fWork, err := os.Create(path.Join(wd, "app.yaml"))
	require.NoError(t, err)
	defer os.Remove(fWork.Name()) //nolint:errcheck

	_, err = fWork.WriteString(data)
	require.NoError(t, err)

	l := loader.File{}

	var c testdata.TestConfig
	require.NoError(t, l.Load("app", &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      112.34,
		Host:        "example.com",
		InnerStruct: testdata.InnerStruct{Str: "test_me"},
	}, c)

	os.Setenv(loader.EnvConfigFile, fEtc.Name()) //nolint:errcheck
	defer os.Unsetenv(loader.EnvConfigFile)      //nolint:errcheck

	require.NoError(t, l.Load("app", &c))

	assert.Equal(t, testdata.TestConfig{
		Salary:      999.9,
		Host:        "exampleetc.com",
		InnerStruct: testdata.InnerStruct{Str: "test_me_ETC"},
	}, c)
}

func TestFile_Nothing(t *testing.T) {
	l := loader.File{}
	var c testdata.TestConfig
	require.Error(t, l.Load("appForNothing", &c))
}

func TestFile_LoadWithUntaggedStruct(t *testing.T) {
	data := `name: Holland
salary: 112.34
host: example.com
innerstruct:
  str: test_me`

	f, err := os.CreateTemp("", "*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name()) //nolint:errcheck

	_, err = f.WriteString(data)
	require.NoError(t, err)

	os.Setenv(loader.EnvConfigFile, f.Name()) //nolint:errcheck
	defer os.Unsetenv(loader.EnvConfigFile)   //nolint:errcheck

	l := loader.File{}

	var c testdata.UntaggedTestConfig
	require.NoError(t, l.Load("a", &c))

	assert.Equal(t, testdata.UntaggedTestConfig{
		Name:        "Holland",
		Salary:      112.34,
		Host:        "example.com",
		InnerStruct: testdata.UntaggedInnerStruct{Str: "test_me"},
	}, c)
}
