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
}

type adminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{adminRepo}
}

func (s *adminService) CreateUser(username, email, password, fullName, roleName string) (models.User, error) {
	// 1. Hash Password
	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	// 2. Find Role ID
	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil {
		return models.User{}, errors.New("role not found")
	}

	newUser := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		RoleID:       role.ID,
		IsActive:     true,
	}

	return s.adminRepo.CreateUser(newUser)
}

func (s *adminService) AssignRole(userID uuid.UUID, roleName string) error {
	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil {
		return errors.New("role not found")
	}
	return s.adminRepo.UpdateUserRole(userID, role.ID)
}