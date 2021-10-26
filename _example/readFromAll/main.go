package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2"
)

// Config struct detailing all project parameters.
type Config struct {
	AppPort int `cfg:"port" env:"PORT" cmd:"port,p" default:"9057"`
	// application specific vault
	DBSchema     string `cfg:"pgxschema"          env:"SCHEMA"             secret:"schema,loggable"      default:"transaction"`
	DBDataSource string `cfg:"pgxdbdatasource"    env:"DBDATASOURCE"       secret:"datasource" loggable:"true"`
	DBType       string `cfg:"dbtype"             env:"DBTYPE"             secret:"dbtype,loggable"      default:"pgx" loggable:"false"`

	SuperSecret SuperSecret `secret:"super-secret,loggable"`
	Test        string      `secret:"test,loggable"`

	// generic vault secrets
	Keycloack Keycloack
	// env automatically compare uppercase
	Migrate Migrate `secret:"migrations,loggable" env:"migrations"`
}

type SuperSecret struct {
	Topsecret string `secret:"topsecret,loggable" default:""`
}

type Migrate struct {
	DBDatasource string `env:"DBDATASOURCE" secret:"dbdatasource,loggable" default:""`
	GetENV       string `env:"TEST_ENV" loggable:"true" default:"X"`
}

type Keycloack struct {
	SSOBaseURL     string `cfg:"ssobaseurl"       secret:"ssobaseurl,loggable"`
	SSOPublicKeyID string `cfg:"ssopubkeyid"      secret:"ssopubkeyid,loggable"`
	SSORealm       string `cfg:"ssorealm"         secret:"ssorealm,loggable"`
	CacheRetention string `cfg:"cacheretention"   default:"30m"`
}

func main() {
	// pretty logging
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out: os.Stderr,
			FormatTimestamp: func(i interface{}) string {
				parse, _ := time.Parse(time.RFC3339, i.(string))

				return parse.Format("2006-01-02 15:04:05")
			},
		}).With().Timestamp().Logger()

	// run igconfig

	// get logger context for config
	logConfig := log.With().Str("component", "config").Logger()
	ctx := logConfig.WithContext(context.Background())

	var conf Config
	if err := igconfig.LoadConfigWithContext(ctx, "test", &conf); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("unable to load configuration settings.")
	}

	// print values
	log.Ctx(ctx).Info().
		EmbedObject(igconfig.Printer{Value: conf}).Msg("loaded config")

	log.Ctx(ctx).Info().Msg(conf.Test)
}
