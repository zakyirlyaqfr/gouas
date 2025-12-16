package middleware

import (
	"errors"
	"gouas/database" // Import DB Instance
	"gouas/app/models"
	"gouas/helper"
	"strings"
)

type AuthResult struct {
	UserID      string
	Role        string
	Permissions []string
}

// CheckAuth memvalidasi token dan mengecek status Whitelist di DB
func CheckAuth(authHeader string) (*AuthResult, error) {
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid token format")
	}

	// 1. Validasi Signature (Stateless)
	claims, err := helper.ValidateJWT(parts[1])
	if err != nil {
		return nil, err
	}

	// 2. Validasi ke Database (Stateful)
	// Cek apakah TokenID di JWT == CurrentAccessTokenID di DB
	var user models.User
	result := database.DB.Select("current_access_token_id").First(&user, "id = ?", claims.UserID)
	
	if result.Error != nil {
		return nil, errors.New("user not found")
	}

	// Jika ID di DB null atau beda dengan ID di token -> TOLAK
	if user.CurrentAccessTokenID == nil || *user.CurrentAccessTokenID != claims.TokenID {
		return nil, errors.New("token has been revoked (logged out or refreshed)")
	}

	return &AuthResult{
		UserID:      claims.UserID.String(),
		Role:        claims.Role,
		Permissions: claims.Permissions,
	}, nil
}

func HasPermission(userPerms []string, requiredPerm string) bool {
	for _, p := range userPerms {
		if p == requiredPerm {
			return true
		}
	}
	return false
}