package igconfig

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/internal"
	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
)

var DefaultLoaders = [...]loader.Loader{
	loader.Default{},
	loader.Consul{},
	loader.Vault{},
	loader.Reader{},
	loader.Env{},
	loader.Flags{},
}

// LoadConfig loads a configuration struct from a fileName, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
// This is a convenience function encapsulating all individual loaders specified in DefaultLoaders.
func LoadConfig(appName string, c interface{}) error {
	return LoadWithLoaders(appName, c, DefaultLoaders[:]...)
}

// LoadWithLoaders uses provided Loader's to fill 'configStruct'.
func LoadWithLoaders(appName string, configStruct interface{}, loaders ...loader.Loader) error {
	for _, configLoader := range loaders {
		err := configLoader.Load(appName, configStruct)
		if err == nil {
			continue
		}

		if errors.Is(err, loader.ErrNoClient) {
			log.Info().
				Str("loader", fmt.Sprintf("%T", configLoader)).
				Msg("no client available, skipping")

			continue
		}

		if internal.IsLocalNetworkError(err) {
			log.Warn().
				Str("loader", fmt.Sprintf("%T", configLoader)).
				Msg("local server is not available, skipping")

			continue
		}

		return fmt.Errorf("%T: %w", configLoader, err)
	}

	return nil
}
