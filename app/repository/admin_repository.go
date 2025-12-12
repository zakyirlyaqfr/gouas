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
	FindAllUsers() ([]models.User, error)
	FindUserByID(id uuid.UUID) (*models.User, error)
	UpdateUser(user models.User) error
	DeleteUser(id uuid.UUID) error
	// NEW: Helper untuk auto-create profile
	CreateStudentProfile(student models.Student) error
	CreateLecturerProfile(lecturer models.Lecturer) error
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

func (r *adminRepository) FindAllUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

func (r *adminRepository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").First(&user, "id = ?", id).Error
	return &user, err
}

func (r *adminRepository) UpdateUser(user models.User) error {
	return r.db.Save(&user).Error
}

func (r *adminRepository) DeleteUser(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// --- NEW IMPLEMENTATION ---
func (r *adminRepository) CreateStudentProfile(student models.Student) error {
	return r.db.Create(&student).Error
}

func (r *adminRepository) CreateLecturerProfile(lecturer models.Lecturer) error {
	return r.db.Create(&lecturer).Error
}
