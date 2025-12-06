package service

import (
	"errors"
	"gouas/app/model"
	"gouas/app/repository"
	"gouas/utils"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(username, email, password, fullName string, roleID uuid.UUID) (*model.User, error)
	Login(username, password string) (string, *model.User, error)
	GetProfile(userID uuid.UUID) (*model.User, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(username, email, password, fullName string, roleID uuid.UUID) (*model.User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		RoleID:       roleID,
		IsActive:     true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(username, password string) (string, *model.User, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", nil, errors.New("invalid username or password")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return "", nil, errors.New("invalid username or password")
	}

	token, err := utils.GenerateToken(user.ID, user.RoleID)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *authService) GetProfile(userID uuid.UUID) (*model.User, error) {
	return s.repo.FindByID(userID)
}