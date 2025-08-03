package main

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const KMFG_TINY_WEB_PORT = 30109
const PASETO_COOKIE_NAME = "tiny.kmfg.ui.auth"

func setupWeb() {
	config := &ServerConfig{
		DefaultPort:  KMFG_TINY_WEB_PORT,
		PortEnvVar:   "KMFG_TINY_WEB_PORT",
		Logger:       &WEB_LOGGER,
		EnableTLS:    os.Getenv("KMFG_TINY_API_TLS") != "false",
		UseTemplates: true,
		TemplateDir:  "./views",
		TemplateExt:  ".html",
	}

	config.startServer(setupWebRoutes)
}

func setupWebRoutes(web *fiber.App) *fiber.App {
	web.Use(webMiddleware)

	signatureGen := setupStaticRouting(web)

	web.Get("/robots.txt", func(c *fiber.Ctx) error {
		return c.SendString("User-Agent: *\nDisallow: /\n")
	})
	web.Get("/", func(c *fiber.Ctx) error {
		return index(c, signatureGen)
	})

	return web
}

func webMiddleware(c *fiber.Ctx) error {
	c.Locals("startTime", time.Now())

	if !strings.HasPrefix(c.Path(), "/static/") {
		logContext(WEB_LOGGER.Info(), c).
			Str("requestMethod", c.Method()).
			Msg("")
	}

	paseto := c.Cookies(PASETO_COOKIE_NAME, "")

	if paseto != "" {
		userPayload, err := FromPaseto(paseto)
		if err != nil {
			WEB_LOGGER.Error().Err(err).Msg("Failed to load user payload from paseto.")
		} else {
			c.Locals("authenticatedUserPayload", userPayload)
		}
	}

	return c.Next()
}

func setupStaticRouting(web *fiber.App) *StaticSignature {
	signatureGen := &StaticSignature{
		Logger: &WEB_LOGGER,
	}
	signatureGen.startSignatureGeneration()

	staticRouter := web.Group("/static")
	staticRouter.Use(func(c *fiber.Ctx) error {
		return staticMiddleware(c, signatureGen)
	})
	staticRouter.Static("/", "./static")

	return signatureGen
}

func staticMiddleware(c *fiber.Ctx, signatureGen *StaticSignature) error {
	signature := c.Query("signature", "")
	sigIsNotOkay := signature == "" || !signatureGen.DoSignaturesMatch(signature)
	if sigIsNotOkay {
		logContext(WEB_LOGGER.Warn(), c).
			Str("signature", signature).
			Msg("Attempt to access static file failed.")

		return c.SendStatus(fiber.ErrForbidden.Code)
	} else {
		logContext(WEB_LOGGER.Info(), c).
			Msg("Static file accessed.")
	}
	return c.Next()
}
