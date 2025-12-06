package middleware

import (
	"gouas/app/model"
	"gouas/database"
	"gouas/helper"
	"os"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Protected: Validasi Token JWT
func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: Invalid or missing token")
		},
	})
}

// PermissionCheck: Validasi Hak Akses (RBAC)
func PermissionCheck(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil claims dari token
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userIDStr := claims["user_id"].(string)

		// 2. Cek DB untuk permission
		var user model.User
		if err := database.DB.Preload("Role.Permissions").Where("id = ?", userIDStr).First(&user).Error; err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "User not found")
		}

		// 3. Loop permission
		for _, perm := range user.Role.Permissions {
			if perm.Name == requiredPermission {
				return c.Next()
			}
		}

		return helper.ErrorResponse(c, fiber.StatusForbidden, "Forbidden: Access denied")
	}
}