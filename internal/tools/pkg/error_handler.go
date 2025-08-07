// Package pkg provides error handling utilities
package pkg

import "github.com/gofiber/fiber/v3"

type AppResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Err     string      `json:"err,omitempty"`
}

// NewAppResponse status 1 err 0 success
func NewAppResponse(ctx fiber.Ctx, httpStatus int, status int, message string, err string, data interface{}) error {
	return ctx.Status(httpStatus).JSON(AppResponse{
		Status:  status,
		Message: message,
		Err:     err,
		Data:    data,
	})
}
