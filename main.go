package main

import (
	"os"
	"os/signal"
	"sync"

	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var (
	DB_LOGGER         zerolog.Logger
	API_LOGGER        zerolog.Logger
	WEB_LOGGER        zerolog.Logger
	TRUSTED_PROXIES   []string
	ADMIN_NEEDS_SETUP bool = false
	adminSetupMutex   sync.RWMutex
)

func main() {
	setupLog()

	log.Info().Int("pid", os.Getpid()).Msg("tiny.kmfg is starting...")

	setupDb()
	checkForAdmin()
	setupTrustedProxies()
	setupTLS()
	setupApi()
	validateXChaCha()
	setupWeb()

	log.Info().Int("pid", os.Getpid()).Msg("All tiny.kmfg services are ready.")

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

func setupTrustedProxies() {
	TRUSTED_PROXIES = ParseTrustedProxies()
}

func setupTLS() {
	err := ensureTLSCertificates()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to ensure TLS certificates")
	}
}

func checkForAdmin() {
	var dbErr error
	var adminCount int64
	dbErr = db.Model(&User{}).Where("is_admin = ?", true).Count(&adminCount).Error

	if dbErr != nil && dbErr != gorm.ErrRecordNotFound {
		WEB_LOGGER.Fatal().Err(dbErr).Msg("Failed to check if any admin users exist.")
	}

	adminSetupMutex.Lock()
	ADMIN_NEEDS_SETUP = dbErr == nil && adminCount > 0
	adminSetupMutex.Unlock()
}
