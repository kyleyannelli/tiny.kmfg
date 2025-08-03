package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog"
)

type ServerConfig struct {
	Port         int
	DefaultPort  int
	PortEnvVar   string
	Logger       *zerolog.Logger
	EnableTLS    bool
	UseTemplates bool
	TemplateDir  string
	TemplateExt  string
}

func (sc *ServerConfig) createFiberApp() *fiber.App {
	config := fiber.Config{
		DisableStartupMessage: true,
	}

	if sc.UseTemplates {
		engine := html.New(sc.TemplateDir, sc.TemplateExt)
		config.Views = engine
	}

	if len(TRUSTED_PROXIES) > 0 {
		sc.Logger.Info().Any("trustedProxies", TRUSTED_PROXIES).Msg("Using trusted proxies.")
		config.EnableTrustedProxyCheck = true
		config.TrustedProxies = TRUSTED_PROXIES
		config.ProxyHeader = fiber.HeaderXForwardedFor
	}

	return fiber.New(config)
}

func (sc *ServerConfig) getPort() int {
	serverPort := sc.DefaultPort
	if sc.PortEnvVar != "" {
		if serverPortStr := os.Getenv(sc.PortEnvVar); serverPortStr != "" {
			if port, err := strconv.Atoi(serverPortStr); err != nil {
				sc.Logger.Fatal().Str("port", serverPortStr).Msg("Cannot convert given port to an int!")
			} else {
				serverPort = port
			}
		}
	}
	return serverPort
}

func (sc *ServerConfig) listen(app *fiber.App, port int) {
	portFmt := fmt.Sprintf(":%d", port)

	if sc.EnableTLS {
		certFile, keyFile := getTLSCertPaths()
		err := app.ListenTLS(portFmt, certFile, keyFile)
		if err != nil {
			sc.Logger.Fatal().Int("port", port).Err(err).Msg("Could not start TLS server.")
		}
	} else {
		err := app.Listen(portFmt)
		if err != nil {
			sc.Logger.Fatal().Int("port", port).Err(err).Msg("Could not start server on specified port.")
		}
	}
}

func (sc *ServerConfig) startServer(setupRoutes func(*fiber.App) *fiber.App) {
	port := sc.getPort()
	app := sc.createFiberApp()
	app = setupRoutes(app)

	go sc.listen(app, port)

	sc.Logger.Info().
		Int("port", port).
		Bool("https", sc.EnableTLS).
		Msg("Started server")
}
