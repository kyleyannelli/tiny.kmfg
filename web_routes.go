package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"kmfg.dev/tiny/auth"
)

var isSetup = false

func index(c *fiber.Ctx, sigGen *StaticSignature) error {
	var dbErr error
	if !isSetup {
		var user User
		dbErr = db.Where("is_admin = ?", true).First(&user).Error
		isSetup = dbErr != gorm.ErrRecordNotFound && dbErr == nil
	}

	if dbErr != nil && dbErr != gorm.ErrRecordNotFound {
		return sigGen.RenderWithDuration(c, "error", fiber.Map{}, "layouts/main")
	}

	var userPayload *auth.UserPayload
	if c.Locals("authenticatedUserPayload") != nil {
		userPayload = c.Locals("authenticatedUserPayload").(*auth.UserPayload)
	}

	if userPayload != nil {
		return dashboard(c, sigGen, userPayload)
	}

	return sigGen.RenderWithDuration(c, "index", fiber.Map{
		"InitSetup": !isSetup,
	}, "layouts/main")
}

func dashboard(c *fiber.Ctx, sigGen *StaticSignature, userPayload *auth.UserPayload) error {
	return sigGen.RenderWithDuration(c, "index", fiber.Map{
		"IsSetup":     isSetup,
		"UserPayload": userPayload,
	}, "layouts/main")
}
