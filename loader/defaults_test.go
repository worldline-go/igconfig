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
	const funcName = "TestDefaultsValues"
	const host = "test"
	const port = 8080

	var c = testdata.TestConfig{
		Host: host,
	}

	require.NoError(t, (loader.Default{}).Load("", &c))

	if c.Name != "Jan" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Jan")
	}
	if c.Age != 18 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 18)
	}
	if c.Salary != 2000.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 2000.0)
	}
	if c.Host != host {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, host)
	}
	if c.Port != port {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, port)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
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
