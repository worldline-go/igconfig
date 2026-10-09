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

func TestNameWithPrefix(t *testing.T) {
	tests := []struct {
		prefix string
		name   string
		want   string
	}{
		{prefix: "", name: "app", want: "app"},
		{prefix: "team", name: "app", want: "team/app"},
		{prefix: "/team/", name: "/app/", want: "team/app"},
		{prefix: "a/b", name: "app", want: "a/b/app"},
	}

	for _, tt := range tests {
		t.Setenv(igconfig.ConfigNamePrefixEnv, tt.prefix)
		require.Equal(t, tt.want, igconfig.NameWithPrefix(tt.name))
	}
}
