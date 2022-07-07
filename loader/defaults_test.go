package loader_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/igconfig/testdata"
)

type OtherInner struct {
	Another string `default:"another"`
}

type Inner struct {
	OtherVal string `default:"other_default"`
	OtherInner
	unexported string   `default:"value"` //nolint:unused
	Slice      []string `default:"1,3,4,5"`
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
		Slice:   []string{"1", "2", "5", "6"},
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
				OtherInner: OtherInner{Another: "another"},
				Slice:      []string{"1", "3", "4", "5"},
			},
		}, v)
}
