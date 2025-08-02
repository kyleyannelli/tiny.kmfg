package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func logContext(event *zerolog.Event, c *fiber.Ctx) *zerolog.Event {
	return event.
		Str("uri", c.OriginalURL()).
		Str("ipAddress", c.IP()).
		Str("referer", string(c.Request().Header.Referer())).
		Str("userAgent", string(c.Request().Header.UserAgent()))
}
