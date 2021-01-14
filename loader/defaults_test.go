package loader_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

type OtherInner struct {
	Another string `default:"another"`
}

type Inner struct {
	OtherVal string `default:"other_default"`
	OtherInner
	unexported string `default:"value"`
}

type StructWithInner struct {
	Val int `default:"1"`
	Inner
}

func TestDefaultsValues(t *testing.T) {
	const host = "test"
	const port = 8080

	var c = testdata.TestConfig{
		Host: host,
	}

	require.NoError(t, (loader.Default{}).Load("", &c))

	assert.Equal(t, testdata.TestConfig{
		Name:    "Jan",
		Age:     18,
		Salary:  2000.0,
		Host:    host,
		Address: "localhost",
		Port:    port,
		Secure:  false,
		InnerStruct: testdata.InnerStruct{
			Str:  "val",
			Time: testdata.ParsedTime,
		},
	}, c)
}

func TestDefault_WithInnerStruct(t *testing.T) {
	var v StructWithInner

	require.NoError(t, (loader.Default{}).Load("", &v))

	assert.Equal(t,
		StructWithInner{Val: 1,
			Inner: Inner{OtherVal: "other_default",
				OtherInner: OtherInner{Another: "another"}},
		}, v)
}
