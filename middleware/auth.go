package middleware

import (
	"errors"
	"gouas/helper"
	"strings"
)

// AuthResult menyimpan hasil validasi middleware
type AuthResult struct {
	UserID      string
	Role        string
	Permissions []string
}

// CheckAuth memvalidasi token Bearer string
func CheckAuth(authHeader string) (*AuthResult, error) {
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid token format")
	}

	claims, err := helper.ValidateJWT(parts[1])
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		UserID:      claims.UserID.String(),
		Role:        claims.Role,
		Permissions: claims.Permissions,
	}, nil
}

// HasPermission mengecek apakah user memiliki permission tertentu
func HasPermission(userPerms []string, requiredPerm string) bool {
	for _, p := range userPerms {
		if p == requiredPerm {
			return true
		}
	}
	return false
}