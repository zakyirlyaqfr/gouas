package middleware

import (
	"gouas/app/model"
	"gouas/database"
	"gouas/helper"
	"os"

	jwtware "github.com/gofiber/contrib/jwt" // <-- Pakai Library Contrib (Support v5)
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected melindungi route dengan JWT
func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		// Konfigurasi SigningKey beda sedikit di library contrib
		SigningKey: jwtware.SigningKey{
			Key: []byte(os.Getenv("JWT_SECRET")),
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: "+err.Error())
		},
	})
}

// PermissionCheck mengecek apakah user memiliki permission tertentu
func PermissionCheck(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil token dari Locals (sekarang sudah pasti v5 karena pakai contrib/jwt)
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userIDStr := claims["user_id"].(string)

		// 2. Cek DB untuk permission
		var user model.User
		// Preload Role & Permissions
		if err := database.DB.Preload("Role.Permissions").Where("id = ?", userIDStr).First(&user).Error; err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "User data not found")
		}

		// 3. Loop permission
		hasPermission := false
		for _, perm := range user.Role.Permissions {
			if perm.Name == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Forbidden: You don't have permission to access this resource")
		}

		return c.Next()
	}
}