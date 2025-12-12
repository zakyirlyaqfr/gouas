package repository

import (
	"gouas/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminRepository interface {
	CreateUser(user models.User) (models.User, error)
	FindRoleByName(name string) (models.Role, error)
	UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) error
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db}
}

func (r *adminRepository) CreateUser(user models.User) (models.User, error) {
	err := r.db.Create(&user).Error
	return user, err
}

func (r *adminRepository) FindRoleByName(name string) (models.Role, error) {
	var role models.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return role, err
}

func (r *adminRepository) UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}