package helper

import (
	"errors"
	"gouas/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}

// GenerateTokens membuat Access Token (15 Menit) dan Refresh Token (30 Menit)
func GenerateTokens(userID uuid.UUID, roleName string, permissions []string) (string, string, error) {
	// 1. Create Access Token (Short Lived - 15 Menit)
	accessClaims := JWTClaims{
		UserID:      userID,
		Role:        roleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // UBAH KE 15 MENIT
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gouas-backend",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(config.GetEnv("JWT_SECRET", "secret")))
	if err != nil {
		return "", "", err
	}

	// 2. Create Refresh Token (Long Lived - 30 Menit)
	// Refresh token biasanya cukup membawa identitas user saja (UserID)
	refreshClaims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // UBAH KE 30 MENIT
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gouas-backend",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.GetEnv("JWT_SECRET", "secret")))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetEnv("JWT_SECRET", "secret")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}