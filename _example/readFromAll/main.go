package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/worldline-go/igconfig"
	"github.com/worldline-go/igconfig/loader"
)

// Config struct detailing all project parameters.
type Config struct {
	AppPort int `cfg:"appPort" env:"PORT" cmd:"port,p" default:"9057"`
	// application specific vault
	DBSchema     string `cfg:"dbSchema"     env:"DBSCHEMA"     secret:"dbSchema,loggable" default:"transaction"`
	DBDataSource string `cfg:"dbDataSource" env:"DBDATASOURCE" secret:"dbDataSource"      loggable:"true"`
	DBType       string `cfg:"dbType"       env:"DBTYPE"       secret:"dbType,loggable"   default:"pgx" loggable:"false"`

	SuperSecret SuperSecret `cfg:"super_secret"`
	Test        string      `cfg:"test-value"`

	// generic vault secrets
	Keycloack Keycloack
	// env automatically compare uppercase
	Migrate Migrate `cfg:"migrations" secret:"migrations,loggable" env:"migrations"`

	Abc interface{}

	LoadTest LoadTest `cfg:"loadtest" secret:"loadtest,loggable"`
}

type LoadTest struct {
	Data string `cfg:"data" secret:"data,loggable"`
}

type SuperSecret struct {
	Topsecret string `secret:"topSecret,loggable" default:""`
}

type Migrate struct {
	DBDatasource string `env:"DBDATASOURCE" secret:"dbDataSource,loggable" default:""`
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
		}).With().Timestamp().Caller().Logger()

	// run igconfig

	// get logger context for config
	logConfig := log.With().Str("component", "config").Logger()
	ctx := logConfig.WithContext(context.Background())

	// ctx := context.Background()

	loader.VaultSecretAdditionalPaths = append(
		loader.VaultSecretAdditionalPaths,
		loader.AdditionalPath{Map: "loadtest", Name: "loadtest"},
	)

	var conf Config
	if err := igconfig.LoadConfigWithContext(ctx, "test2", &conf); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("unable to load configuration settings.")
	}

	// print values
	log.Ctx(ctx).Info().
		EmbedObject(igconfig.Printer{Value: conf}).Msg("loaded config")

	log.Ctx(ctx).Info().Msg(conf.Test)
}
