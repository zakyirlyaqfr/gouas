package service

import (
	"errors"
	"gouas/app/repository"
	"gouas/helper"
)

type AuthService interface {
	Login(username, password string) (string, error)
}

type authService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{authRepo}
}

func (s *authService) Login(username, password string) (string, error) {
	// 1. Cari user
	user, err := s.authRepo.FindByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 2. Cek Password
	if !helper.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	// 3. Cek Active
	if !user.IsActive {
		return "", errors.New("user is inactive")
	}

	// 4. Collect Permissions
	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	// 5. Generate Token
	token, err := helper.GenerateJWT(user.ID, user.Role.Name, permissions)
	if err != nil {
		return "", err
	}

	return token, nil
}