package service

import (
	"errors"
	"gouas/app/repository"
	"gouas/helper"

	"github.com/google/uuid"
)

type AuthService interface {
	Login(username, password string) (string, string, error)
	Refresh(refreshToken string) (string, error) // Refresh Access Token
	Logout(userID uuid.UUID) error               // Logout
}

type authService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{authRepo}
}

func (s *authService) Login(username, password string) (string, string, error) {
	// 1. Validasi User
	user, err := s.authRepo.FindByUsername(username)
	if err != nil { return "", "", errors.New("invalid credentials") }
	if !helper.CheckPasswordHash(password, user.PasswordHash) { return "", "", errors.New("invalid credentials") }
	if !user.IsActive { return "", "", errors.New("user is inactive") }

	// 2. Generate ID Baru untuk Token
	newAccessID := uuid.New()
	newRefreshID := uuid.New()

	// 3. Simpan ID ke Database (Whitelist)
	err = s.authRepo.UpdateTokenIDs(user.ID, &newAccessID, &newRefreshID)
	if err != nil { return "", "", err }

	// 4. Generate Token JWT
	permissions := []string{}
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	accessToken, _ := helper.GenerateAccessToken(user.ID, user.Role.Name, permissions, newAccessID)
	refreshToken, _ := helper.GenerateRefreshToken(user.ID, newRefreshID)

	return accessToken, refreshToken, nil
}

func (s *authService) Refresh(refreshToken string) (string, error) {
	// 1. Validasi Signature Refresh Token
	claims, err := helper.ValidateJWT(refreshToken)
	if err != nil { return "", errors.New("invalid refresh token") }

	// 2. Cek DB: Apakah ID Refresh Token ini cocok dengan yang ada di DB?
	user, err := s.authRepo.FindByID(claims.UserID)
	if err != nil { return "", errors.New("user not found") }

	if user.CurrentRefreshTokenID == nil || *user.CurrentRefreshTokenID != claims.TokenID {
		return "", errors.New("refresh token invalid or revoked")
	}

	// 3. Generate ID Access Token BARU (Ini yang membuat access token lama mati!)
	newAccessID := uuid.New()

	// Update DB: Ganti Access ID, tapi biarkan Refresh ID (karena refresh token berlaku 24 jam)
	// Kita kirim nil untuk refreshID di repo agar tidak diupdate/dihapus
	// Tapi logika repo kita tadi perlu disesuaikan sedikit, atau kita bisa kirim CurrentRefreshTokenID lagi
	err = s.authRepo.UpdateTokenIDs(user.ID, &newAccessID, user.CurrentRefreshTokenID) 
	if err != nil { return "", err }

	// 4. Buat Access Token Baru
	permissions := []string{}
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}
	
	newAccessToken, _ := helper.GenerateAccessToken(user.ID, user.Role.Name, permissions, newAccessID)
	
	return newAccessToken, nil
}

func (s *authService) Logout(userID uuid.UUID) error {
	// Set semua ID token jadi NULL -> Semua token hangus
	return s.authRepo.UpdateTokenIDs(userID, nil, nil)
}