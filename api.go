package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const KMFG_TINY_API_PORT = 30108

func setupApi() {
	serverPort := KMFG_TINY_API_PORT
	serverPortStr := os.Getenv("KMFG_TINY_API_PORT")
	var err error
	if serverPortStr != "" {
		serverPort, err = strconv.Atoi(serverPortStr)
	}
	if err != nil {
		API_LOGGER.Fatal().Str("port", serverPortStr).Msg("Cannot convert given port to an int!")
	}
	api := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	api.Get("/:shortCode", redirectURL)
	go listenAloneApi(api, serverPort)

	API_LOGGER.Info().Int("port", serverPort).Msg("Started API server")
}

func listenAloneApi(api *fiber.App, serverPort int) {
	err := api.Listen(fmt.Sprintf(":%d", serverPort))
	if err != nil {
		API_LOGGER.Fatal().Int("port", serverPort).Err(err).Msg("Could not listen on specified port.")
	}
}

func redirectURL(c *fiber.Ctx) error {
	shortCode := c.Params("shortCode")

	if shortCode == "" {
		logContext(API_LOGGER.Warn(), c).
			Msg("No short code found...")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Short code is required",
		})
	}

	var tinyUrl TinyUrl
	res := db.Where("short_code = ?", shortCode).First(&tinyUrl)

	if res.Error == gorm.ErrRecordNotFound {
		logContext(API_LOGGER.Info(), c).
			Str("shortCode", shortCode).
			Err(res.Error).
			Msg("Short URL not found in database")
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Short URL not found",
		})
	} else if res.Error != nil {
		logContext(API_LOGGER.Error(), c).
			Str("shortCode", shortCode).
			Err(res.Error).
			Msg("Database error while fetching short URL")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal error",
		})
	}

	logContext(API_LOGGER.Info(), c).Str("shortCode", shortCode).Msg("")

	return c.Redirect(tinyUrl.TrueUrl)
}
