package main

import (
	"os"
	"os/signal"

	"syscall"

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down all services...")
}

func setupLog() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	DB_LOGGER = log.With().Str("service", "database").Logger()
	API_LOGGER = log.With().Str("service", "api").Logger()
	WEB_LOGGER = log.With().Str("service", "web").Logger()
}
