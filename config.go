package igconfig

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
)

var DefaultLoaders = [...]loader.Loader{
	loader.Default{},
	loader.Consul{},
	loader.Vault{},
	loader.Env{},
	loader.Flags{},
}

// LoadConfig loads a configuration struct from a fileName, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
// This is a convenience function encapsulating all individual functions above.
//
// Note: this does not load configuration from any kind of files.
// If this is required - use loader.Reader to load data from file.
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

		// Check if error is network one.
		var netErr *net.OpError
		if errors.As(err, &netErr) {
			// Check if network error is well-known "Connection Refused"
			var sErr *os.SyscallError
			if !errors.As(err, &sErr) || !errors.Is(sErr.Err, syscall.ECONNREFUSED) {
				// If it is not a connection refused - return it.
				return err
			}

			// If host is 127.0.0.1 - it means that no hostname was provided in environment.
			// Please use "localhost" if you want to receive an error instead.
			if strings.HasPrefix(netErr.Addr.String(), "127.0.0.1") {
				log.Warn().
					Str("loader", fmt.Sprintf("%T", configLoader)).
					Msg("local server is not available, skipping")

				continue
			}
		}

		return fmt.Errorf("%T: %w", configLoader, err)
	}

	return nil
}
