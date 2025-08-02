package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	DB_LOGGER  zerolog.Logger
	API_LOGGER zerolog.Logger
	WEB_LOGGER zerolog.Logger
)

func main() {
	setupLog()
	setupDb()
	setupApi()
	setupWeb()

	time.Sleep(60 * time.Minute)
}

func setupLog() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	DB_LOGGER = log.With().Str("service", "database").Logger()
	API_LOGGER = log.With().Str("service", "api").Logger()
	WEB_LOGGER = log.With().Str("service", "web").Logger()
}
