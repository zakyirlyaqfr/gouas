package service

import (
	"errors"
	"gouas/app/models"
	"gouas/app/repository"
	"gouas/helper"

	"github.com/google/uuid"
)

type AdminService interface {
	CreateUser(username, email, password, fullName, roleName string) (models.User, error)
	AssignRole(userID uuid.UUID, roleName string) error
	// New
	GetAllUsers() ([]models.User, error)
	GetUserDetail(id uuid.UUID) (*models.User, error)
	UpdateUser(id uuid.UUID, fullName, email string) error
	DeleteUser(id uuid.UUID) error
}

type adminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{adminRepo}
}

func (s *adminService) CreateUser(username, email, password, fullName, roleName string) (models.User, error) {
	hashedPassword, err := helper.HashPassword(password)
	if err != nil { return models.User{}, err }
	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil { return models.User{}, errors.New("role not found") }
	newUser := models.User{
		Username: username, Email: email, PasswordHash: hashedPassword, FullName: fullName, RoleID: role.ID, IsActive: true,
	}
	return s.adminRepo.CreateUser(newUser)
}

func (s *adminService) AssignRole(userID uuid.UUID, roleName string) error {
	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil { return errors.New("role not found") }
	return s.adminRepo.UpdateUserRole(userID, role.ID)
}

// --- NEW IMPL ---

func (s *adminService) GetAllUsers() ([]models.User, error) {
	return s.adminRepo.FindAllUsers()
}

func (s *adminService) GetUserDetail(id uuid.UUID) (*models.User, error) {
	return s.adminRepo.FindUserByID(id)
}

func (s *adminService) UpdateUser(id uuid.UUID, fullName, email string) error {
	user, err := s.adminRepo.FindUserByID(id)
	if err != nil { return err }
	
	user.FullName = fullName
	user.Email = email
	return s.adminRepo.UpdateUser(*user)
}

func (s *adminService) DeleteUser(id uuid.UUID) error {
	return s.adminRepo.DeleteUser(id)
}