package loader_test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

func TestFileBadData(t *testing.T) {
	var (
		buf bytes.Buffer
	)
	tests := []struct {
		FileData string
		Error    string
	}{
		{
			FileData: "age=haha",
			Error:    `value for val "Age" not a valid "uint"`,
		},
		{
			FileData: "age",
		},
	}

	var c testdata.TestConfig

	for i, test := range tests {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			l := loader.Reader{}

			buf.WriteString(test.FileData)

			err := l.LoadReader(&buf, &c)
			if test.Error != "" {
				assert.EqualError(t, err, test.Error)
			} else {
				assert.NoError(t, err)
			}

			buf.Reset()
		})
	}
}

func TestFileOverwriteDefault(t *testing.T) {
	var (
		buf      bytes.Buffer
		fileData = "settle_name="
	)
	buf.WriteString(fileData)

	var c testdata.TestConfig

	require.NoError(t, (loader.Default{}).Load("", &c))
	assert.NoError(t, (loader.Reader{}).LoadReader(&buf, &c))

	assert.Equal(t, "", c.Name)
}

func TestFile(t *testing.T) {
	tests := []struct {
		Name   string
		Data   string
		Result testdata.TestConfig
	}{
		{
			Name: "TestFileSimple",
			Data: "age=28\nsalary=1800.00\nsettle_name=Jantje",
			Result: testdata.TestConfig{
				Name:    "Jantje",
				Age:     28,
				Salary:  1800.0,
				Host:    "localhost",
				Address: "localhost",
				Port:    8080,
				Secure:  false,
				Unused:  nil,
			},
		},
		{

			Name: "TestFileComplex",
			Data: "// Age\nage=28\n#Salary\nsalary=1800.00\n\nsettle_name=Jantje\n ## Name of subject ##\nwrong=test\n\n",
			Result: testdata.TestConfig{
				Name:    "Jantje",
				Age:     28,
				Salary:  1800.0,
				Host:    "localhost",
				Address: "localhost",
				Port:    8080,
				Secure:  false,
				Unused:  nil,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteString(test.Data)

			var c testdata.TestConfig
			require.NoError(t, (loader.Default{}).Load("", &c))

			assert.NoError(t, (loader.Reader{}).LoadReader(&buf, &c))

			assert.Equal(t, test.Result, c)
		})
	}
}
