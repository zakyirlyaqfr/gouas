package service

import (
	"errors"
	"gouas/app/repository"
	"gouas/helper"

	// "github.com/google/uuid"
)

type AuthService interface {
	Login(username, password string) (string, string, error) // Return 2 token
	Refresh(refreshToken string) (string, string, error)     // Method Baru
}

type authService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{authRepo}
}

func (s *authService) Login(username, password string) (string, string, error) {
	// 1. Cari user
	user, err := s.authRepo.FindByUsername(username)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// 2. Cek Password
	if !helper.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", errors.New("invalid credentials")
	}

	// 3. Cek Active
	if !user.IsActive {
		return "", "", errors.New("user is inactive")
	}

	// 4. Collect Permissions
	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	// 5. Generate Dual Tokens
	return helper.GenerateTokens(user.ID, user.Role.Name, permissions)
}

func (s *authService) Refresh(refreshToken string) (string, string, error) {
	// 1. Validasi Refresh Token
	claims, err := helper.ValidateJWT(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// 2. Cek User di DB (Untuk memastikan user belum dihapus/inactive & update role terbaru)
	user, err := s.authRepo.FindByID(claims.UserID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	if !user.IsActive {
		return "", "", errors.New("user is inactive")
	}

	// 3. Collect Permissions Terbaru
	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	// 4. Generate Token Baru
	return helper.GenerateTokens(user.ID, user.Role.Name, permissions)
}