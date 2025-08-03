package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

const KMFG_TINY_WEB_PORT = 30109
const PASETO_COOKIE_NAME = "tiny.kmfg.ui.auth"

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

	engine := html.New("./views", ".html")

	var web *fiber.App
	if len(TRUSTED_PROXIES) == 0 {
		web = fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 engine,
		})
	} else {
		WEB_LOGGER.Info().Any("trustedProxies", TRUSTED_PROXIES).Msg("Using trusted proxies.")
		web = fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 engine,

			EnableTrustedProxyCheck: true,
			TrustedProxies:          TRUSTED_PROXIES,
			ProxyHeader:             fiber.HeaderXForwardedFor,
		})
	}

	web.Use(func(c *fiber.Ctx) error {
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
	})

	signatureGen := setupStaticRouting(web)

	web.Get("/", func(c *fiber.Ctx) error {
		return index(c, signatureGen)
	})

	go listenAloneWeb(web, serverPort)

	WEB_LOGGER.Info().Int("port", serverPort).Msg("Started web server")
}

func listenAloneWeb(api *fiber.App, serverPort int) {
	err := api.Listen(fmt.Sprintf(":%d", serverPort))
	if err != nil {
		WEB_LOGGER.Fatal().Int("port", serverPort).Err(err).Msg("Could not listen on specified port.")
	}
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
