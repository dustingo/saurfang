package boardhandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/config"
	"saurfang/internal/models/dashboard"
)

func Handler_LoginRecords(c fiber.Ctx) error {
	var records []dashboard.LoginRecords
	config.DB.Limit(6).Find(&records).Order("last_login desc")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"rows": records,
		},
	})

}
