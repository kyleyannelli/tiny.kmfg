package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const KMFG_TINY_WEB_PORT = 30109

func setupWeb() {
	serverPort := KMFG_TINY_WEB_PORT
	serverPortStr := os.Getenv("KMFG_TINY_WEB_PORT")
	var err error
	if serverPortStr != "" {
		serverPort, err = strconv.Atoi(serverPortStr)
	}
	if err != nil {
		WEB_LOGGER.Fatal().Str("port", serverPortStr).Msg("Cannot convert given port to an int!")
	}

	web := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	go listenAloneWeb(web, serverPort)

	WEB_LOGGER.Info().Int("port", serverPort).Msg("Started web server")
}

func listenAloneWeb(api *fiber.App, serverPort int) {
	err := api.Listen(fmt.Sprintf(":%d", serverPort))
	if err != nil {
		API_LOGGER.Fatal().Int("port", serverPort).Err(err).Msg("Could not listen on specified port.")
	}
}
