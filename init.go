package igconfig

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// init function will prepare logging for proper usage with Fluentd configuration.
func init() { //nolint:gochecknoinits // This is default for all of the services and should be set by default.
	ConfigureZerolog()
}

// ConfigureZerolog can be used to "reset" default global logging configuration for zerolog.
//
// This will update Time format to RFC3339Nano and will set up new global logger.
func ConfigureZerolog() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
	zerolog.DefaultContextLogger = &log.Logger
}
