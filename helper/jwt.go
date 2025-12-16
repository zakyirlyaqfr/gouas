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
	TokenID     uuid.UUID `json:"jti"` // [BARU] ID Unik Token
	jwt.RegisteredClaims
}

// GenerateAccessToken membuat token pendek (15 menit)
func GenerateAccessToken(userID uuid.UUID, roleName string, permissions []string, tokenID uuid.UUID) (string, error) {
	claims := JWTClaims{
		UserID:      userID,
		Role:        roleName,
		Permissions: permissions,
		TokenID:     tokenID, // Simpan ID
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 Menit
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gouas-backend",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetEnv("JWT_SECRET", "secret")))
}

// GenerateRefreshToken membuat token panjang (24 jam)
func GenerateRefreshToken(userID uuid.UUID, tokenID uuid.UUID) (string, error) {
	claims := JWTClaims{
		UserID:  userID,
		TokenID: tokenID, // Simpan ID
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 Jam
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gouas-backend",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetEnv("JWT_SECRET", "secret")))
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