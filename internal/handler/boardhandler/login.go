package boardhandler

import (
	"saurfang/internal/config"
	"saurfang/internal/models/dashboard"

	"github.com/gofiber/fiber/v3"
)

func Handler_LoginRecords(c fiber.Ctx) error {
	var records []dashboard.LoginRecords
	config.DB.Order("last_login desc").Limit(6).Find(&records)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"rows": records,
		},
	})

}
