package igconfig_test

import (
	"os"
	"testing"

	"github.com/worldline-go/igconfig/loader"

	"github.com/worldline-go/igconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/worldline-go/igconfig/testdata"
)

func TestLoadConfig(t *testing.T) {
	os.Clearenv()

	c1 := testdata.BadDefaults{}

	assert.NotNil(t, igconfig.LoadConfig("haha", &c1))

	var c2 testdata.TestConfig

	prevArgs := os.Args
	os.Args = []string{"program", "--age", "haha", "--salary", "nothing"}

	assert.NotNil(t, igconfig.LoadConfig("", &c2))
	os.Args = prevArgs
}

func TestLoadConfig_Skip(t *testing.T) {
	testLoaders := []loader.Loader{
		loader.File{},
	}

	var c testdata.TestConfig

	require.NoError(t, igconfig.LoadWithLoaders("skipApp", &c, testLoaders...))
}
