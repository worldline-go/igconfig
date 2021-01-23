package internal

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

func TestSetValueWarnings(t *testing.T) {
	var c testdata.TestConfig

	tests := []struct {
		Field string
		Val   string
	}{
		{
			Field: "Age",
			Val:   "haha",
		},
		{
			Field: "Port",
			Val:   "haha",
		},
		{
			Field: "Salary",
			Val:   "haha",
		},
		{
			Field: "Unsused",
			Val:   "1.0",
		},
	}

	for i, test := range tests {
		refVal, _ := GetReflectElem(&c)
		assert.NotNil(t, SetStructFieldValue(test.Field, test.Val, refVal), fmt.Sprintf("test #%d", i))
	}
}

func TestSetReflectValueString_UnmarshalText(t *testing.T) {
	var s struct {
		IP net.IP
	}
	strIp := "10.11.12.13"

	assert.NoError(t, SetStructFieldValue("IP", strIp, reflect.ValueOf(&s)))
	assert.Equal(t, strIp, s.IP.String())
}

func TestSetStruct(t *testing.T) {
	testStruct := struct {
		Field struct {
			Test bool
		} `cmd:"struct"`
	}{}

	refVal, _ := GetReflectElem(&testStruct)
	assert.NotNil(t, SetStructFieldValue("Field", `{"test":false`, refVal))
}
