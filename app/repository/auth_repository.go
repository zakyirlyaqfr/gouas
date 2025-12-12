package repository

import (
	"gouas/app/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	FindByUsername(username string) (*models.User, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	// Eager load Role dan Permissions di dalam Role tersebut
	err := r.db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}