package igconfig

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/worldline-go/igconfig/internal"
	"github.com/worldline-go/igconfig/loader"
)

// DefaultLoaders is a list of default loaders to use.
var DefaultLoaders = []loader.Loader{
	&loader.Default{},
	&loader.Consul{},
	&loader.Vault{},
	&loader.File{},
	&loader.Env{},
	&loader.Flags{},
}

// LoadConfig loads a configuration struct from loaders.
func LoadConfig(appName string, c interface{}) error {
	return LoadConfigWithContext(context.Background(), appName, c)
}

// LoadConfigWithContext loads a configuration struct from a fileName, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
// This is a convenience function encapsulating all individual loaders specified in DefaultLoaders.
func LoadConfigWithContext(ctx context.Context, appName string, c interface{}) error {
	return LoadWithLoadersWithContext(ctx, appName, c, DefaultLoaders...)
}

// LoadWithLoaders loads a configuration struct from a loaders.
func LoadWithLoaders(appName string, configStruct interface{}, loaders ...loader.Loader) error {
	return LoadWithLoadersWithContext(context.Background(), appName, configStruct, loaders...)
}

// LoadWithLoadersWithContext uses provided Loader's to fill 'configStruct'.
func LoadWithLoadersWithContext(ctx context.Context, appName string, configStruct interface{}, loaders ...loader.Loader) error {
	for _, configLoader := range loaders {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		err := configLoader.LoadWithContext(ctx, appName, configStruct)
		if err == nil {
			continue
		}

		if errors.Is(err, loader.ErrNoClient) {
			log.Ctx(ctx).Warn().
				Str("loader", fmt.Sprintf("%T", configLoader)).
				Msgf("%v, skipping", err)

			continue
		}

		if internal.IsLocalNetworkError(err) {
			log.Ctx(ctx).Warn().
				Str("loader", fmt.Sprintf("%T", configLoader)).
				Msg("local server is not available, skipping")

			continue
		}

		if errors.Is(err, loader.ErrNoConfFile) {
			log.Ctx(ctx).Warn().
				Str("loader", fmt.Sprintf("%T", configLoader)).
				Msgf("%v, skipping", err)

			continue
		}

		return fmt.Errorf("%T: %w", configLoader, err)
	}

	return nil
}
