package loader_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/igconfig/testdata"
)

func TestEnvValues(t *testing.T) {
	const funcName = "TestEnvValues"

	t.Setenv("NAME", "Jan")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "12345")
	t.Setenv("age", "44")
	t.Setenv("address", "should_not_be_set")
	t.Setenv("INNERSTRUCT_STRING", "hello")
	t.Setenv("SLICE", "3,4,4")

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
