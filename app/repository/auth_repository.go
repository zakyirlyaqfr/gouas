package repository

import (
	"gouas/app/models"

	"github.com/google/uuid" // Pastikan import ini ada
	"gorm.io/gorm"
)

type AuthRepository interface {
	FindByUsername(username string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error) // [BARU] Method untuk Refresh Token
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	// Eager load Role dan Permissions untuk generate token saat Login
	err := r.db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// [BARU] Implementasi FindByID
// Digunakan oleh Service saat melakukan Refresh Token untuk memastikan user masih ada/aktif
func (r *authRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	// Kita perlu Preload Role & Permissions lagi karena token baru akan digenerate
	err := r.db.Preload("Role.Permissions").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}