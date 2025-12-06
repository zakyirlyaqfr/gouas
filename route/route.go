package route

import (
	"gouas/app/repository"
	"gouas/app/service"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Struct Request Body ditaruh disini atau di utils (bebas)
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	RoleID   string `json:"role_id"`
}

func SetupRoutes(app *fiber.App) {
	// 1. Init Dependencies (Manual Injection)
	authRepo := repository.NewAuthRepository()
	authService := service.NewAuthService(authRepo)

	// 2. Group API
	api := app.Group("/api/v1")

	// ================= AUTH ROUTES =================
	
	// POST /api/v1/auth/register
	api.Post("/auth/register", func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		roleUUID, _ := uuid.Parse(req.RoleID)
		user, err := authService.Register(req.Username, req.Email, req.Password, req.FullName, roleUUID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "User registered", user)
	})

	// POST /api/v1/auth/login
	api.Post("/auth/login", func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		token, user, err := authService.Login(req.Username, req.Password)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
		}

		return helper.SuccessResponse(c, "Login success", fiber.Map{
			"token": token,
			"user":  user,
		})
	})

	// GET /api/v1/users/profile (Protected)
	api.Get("/users/profile", middleware.Protected(), func(c *fiber.Ctx) error {
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userIDStr := claims["user_id"].(string)
		userID, _ := uuid.Parse(userIDStr)

		user, err := authService.GetProfile(userID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "User not found")
		}
		return helper.SuccessResponse(c, "User profile", user)
	})
}