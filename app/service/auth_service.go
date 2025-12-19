package service

import (
	"gouas/app/repository"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strings"
)

type AuthService interface {
	Login(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	GetProfile(c *fiber.Ctx) error
}

type authService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{authRepo}
}

func (s *authService) Login(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "Invalid input", nil))
	}

	// 1. Cari user
	user, err := s.authRepo.FindByUsername(input.Username)
	if err != nil {
		return c.Status(401).JSON(helper.APIResponse("error", "Invalid credentials", nil))
	}

	// 2. Cek Password
	if !helper.CheckPasswordHash(input.Password, user.PasswordHash) {
		return c.Status(401).JSON(helper.APIResponse("error", "Invalid credentials", nil))
	}

	// 3. Cek Active
	if !user.IsActive {
		return c.Status(401).JSON(helper.APIResponse("error", "User is inactive", nil))
	}

	// 4. Update Token ID di DB (Stateful)
	newAccessID := uuid.New()
	newRefreshID := uuid.New()
	if err := s.authRepo.UpdateTokenIDs(user.ID, &newAccessID, &newRefreshID); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	// 5. Generate Tokens
	permissions := []string{}
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}
	accessToken, _ := helper.GenerateAccessToken(user.ID, user.Role.Name, permissions, newAccessID)
	refreshToken, _ := helper.GenerateRefreshToken(user.ID, newRefreshID)

	return c.Status(200).JSON(helper.APIResponse("success", "Login successful", fiber.Map{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}))
}

func (s *authService) Refresh(c *fiber.Ctx) error {
	// 1. Ambil token dari header Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(helper.APIResponse("error", "Missing Authorization header", nil))
	}

	// 2. Bersihkan string "Bearer " untuk mendapatkan token murni
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	if tokenString == "" {
		return c.Status(401).JSON(helper.APIResponse("error", "Invalid token format", nil))
	}

	// 3. Validasi Token (Stateless Check)
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return c.Status(401).JSON(helper.APIResponse("error", "Invalid or expired refresh token", nil))
	}

	// 4. Cek DB: Pastikan Refresh Token ini yang terdaftar/aktif (Stateful Check)
	user, err := s.authRepo.FindByID(claims.UserID)
	if err != nil || user.CurrentRefreshTokenID == nil || *user.CurrentRefreshTokenID != claims.TokenID {
		return c.Status(401).JSON(helper.APIResponse("error", "Refresh token has been revoked", nil))
	}

	// 5. Rotasi Access Token ID (Bikin Access Token lama mati)
	newAccessID := uuid.New()
	// Gunakan CurrentRefreshTokenID agar refresh token lama tetap valid sampai 24 jam
	if err := s.authRepo.UpdateTokenIDs(user.ID, &newAccessID, user.CurrentRefreshTokenID); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	// 6. Generate Access Token Baru
	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}
	newAccess, _ := helper.GenerateAccessToken(user.ID, user.Role.Name, permissions, newAccessID)

	return c.Status(200).JSON(helper.APIResponse("success", "Token refreshed", fiber.Map{
		"accessToken":  newAccess,
		"refreshToken": tokenString, // Kembalikan token yang sama karena ID-nya belum berubah
	}))
}

func (s *authService) Logout(c *fiber.Ctx) error {
	authData, err := middleware.CheckAuth(c.Get("Authorization"))
	if err != nil {
		return c.Status(401).JSON(helper.APIResponse("error", "Unauthorized", nil))
	}
	userID, _ := uuid.Parse(authData.UserID)

	if err := s.authRepo.UpdateTokenIDs(userID, nil, nil); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Logged out successfully", nil))
}

func (s *authService) GetProfile(c *fiber.Ctx) error {
	authData, err := middleware.CheckAuth(c.Get("Authorization"))
	if err != nil {
		return c.Status(401).JSON(helper.APIResponse("error", "Unauthorized", nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "User Profile", authData))
}