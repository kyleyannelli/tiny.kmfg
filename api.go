package main

import (
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const KMFG_TINY_API_PORT = 30108

func setupApi() {
	generateRobotsTxt(nil)

	config := &ServerConfig{
		DefaultPort: KMFG_TINY_API_PORT,
		PortEnvVar:  "KMFG_TINY_API_PORT",
		Logger:      &API_LOGGER,
		EnableTLS:   os.Getenv("KMFG_TINY_API_TLS") != "false",
	}

	config.startServer(setupApiRoutes)
}

func setupApiRoutes(api *fiber.App) *fiber.App {
	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusTeapot)
	})
	api.Get("/robots.txt", robots)
	api.Get("/:shortCode", redirectURL)
	return api
}

func robots(c *fiber.Ctx) error {
	_, err := os.Stat(ROBOTS_FILE)
	if err != nil {
		return c.SendString("User-Agent: *\nDisallow: /\n")
	}
	return c.SendFile(ROBOTS_FILE)
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

	visitUrl(&tinyUrl, c)
	logContext(API_LOGGER.Info(), c).Str("shortCode", shortCode).Msg("")

	return c.Redirect(tinyUrl.TrueUrl)
}
