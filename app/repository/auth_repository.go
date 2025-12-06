package repository

import (
	"gouas/app/model"
	"gouas/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uuid.UUID) (*model.User, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository() AuthRepository {
	return &authRepository{db: database.DB}
}

func (r *authRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	// Preload Role & Permissions
	err := r.db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *authRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role.Permissions").Where("id = ?", id).First(&user).Error
	return &user, err
}