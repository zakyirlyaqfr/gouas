package helper

import "github.com/gofiber/fiber/v2"

// APIResponse adalah format standar output JSON
type APIResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse helper untuk respon sukses
func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Code:    fiber.StatusOK,
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// ErrorResponse helper untuk respon error
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Code:    statusCode,
		Status:  "error",
		Message: message,
		Data:    nil,
	})
}