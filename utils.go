package main

import (
	"net"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func logContext(event *zerolog.Event, c *fiber.Ctx) *zerolog.Event {
	return event.
		Str("uri", c.Path()).
		Str("ipAddress", c.IP()).
		Any("origin", c.Context().RemoteIP()).
		Str("referer", string(c.Request().Header.Referer())).
		Str("userAgent", string(c.Request().Header.UserAgent()))
}

func ParseTrustedProxies() []string {
	envVar := os.Getenv("KMFG_TINY_TRUSTED_IPS")
	if envVar == "" {
		return []string{}
	}

	rawIPs := strings.Split(envVar, ",")
	var trustedIPs []string

	for _, rawIP := range rawIPs {
		ip := strings.TrimSpace(rawIP)
		if ip == "" {
			continue
		}

		if net.ParseIP(ip) != nil {
			trustedIPs = append(trustedIPs, ip)
		} else {
			log.Fatal().Str("ipAddress", ip).Msg("Invalid IP given, cannot parse.")
		}
	}

	return trustedIPs
}
