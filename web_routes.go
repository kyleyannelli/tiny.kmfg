package main

import (
	"github.com/gofiber/fiber/v2"
	"kmfg.dev/tiny/auth"
)

func index(c *fiber.Ctx, sigGen *StaticSignature) error {
	var userPayload *auth.UserPayload
	if c.Locals("authenticatedUserPayload") != nil {
		userPayload = c.Locals("authenticatedUserPayload").(*auth.UserPayload)
	}

	adminSetupMutex.RLock()
	if userPayload != nil {
		adminSetupMutex.RUnlock()
		return dashboard(c, sigGen, userPayload)
	} else if ADMIN_NEEDS_SETUP {
		adminSetupMutex.RUnlock()
		return sigGen.RenderWithDuration(c, "login", fiber.Map{}, "layouts/main")
	} else {
		adminSetupMutex.RUnlock()
		return sigGen.RenderWithDuration(c, "first_signup", fiber.Map{}, "layouts/main")
	}
}

func register(c *fiber.Ctx, sigGen *StaticSignature) {
	var userPayload *auth.UserPayload
	if c.Locals("authenticatedUserPayload") != nil {
		userPayload = c.Locals("authenticatedUserPayload").(*auth.UserPayload)
	}
	WEB_LOGGER.Debug().Any("user", userPayload).Msg("Hi")
}

func dashboard(c *fiber.Ctx, sigGen *StaticSignature, userPayload *auth.UserPayload) error {
	return sigGen.RenderWithDuration(c, "dashboard", fiber.Map{
		"UserPayload": userPayload,
	}, "layouts/main")
}
