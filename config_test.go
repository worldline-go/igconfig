package igconfig_test

import (
	"testing"

	"github.com/worldline-go/igconfig/loader"

	"github.com/worldline-go/igconfig"

	"github.com/stretchr/testify/require"

	"github.com/worldline-go/igconfig/testdata"
)

func TestLoadConfig_Skip(t *testing.T) {
	testLoaders := []loader.Loader{
		loader.File{},
	}

	var c testdata.TestConfig

	require.NoError(t, igconfig.LoadWithLoaders("skipApp", &c, testLoaders...))
}
