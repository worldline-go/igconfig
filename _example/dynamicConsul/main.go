package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/worldline-go/igconfig/loader"
)

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

	// get logger context for config
	rootCtx, cancelRootCtx := context.WithCancel(context.Background())

	defer cancelRootCtx()

	log.Logger.Debug().Msg("dynamic listening starting")

	ch, err := loader.Consul{}.DynamicValue(rootCtx, "test")
	if err != nil {
		log.Logger.Debug().Err(err).Msg("uupps")

		return
	}

	// get []byte
	for v := range ch {
		fmt.Printf("%s\n", v)
	}
}
