package igconfig_test

import (
	"os"
	"testing"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
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
