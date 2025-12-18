package test

import (
	"bytes"
	"encoding/json"
	"gouas/app/models"
	"gouas/app/service"
	"gouas/helper"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepo) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepo) UpdateTokenIDs(userID uuid.UUID, accessID *uuid.UUID, refreshID *uuid.UUID) error {
	args := m.Called(userID, accessID, refreshID)
	return args.Error(0)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	authSvc := service.NewAuthService(mockRepo)
	app := fiber.New()
	app.Post("/login", authSvc.Login)

	password := "password123"
	hashed, _ := helper.HashPassword(password)
	userID := uuid.New()

	mockUser := &models.User{
		ID:           userID,
		Username:     "mahasiswa1",
		PasswordHash: hashed,
		IsActive:     true,
		Role: models.Role{
			Name: "Mahasiswa",
			Permissions: []models.Permission{
				{Name: "achievement:create"},
			},
		},
	}

	mockRepo.On("FindByUsername", "mahasiswa1").Return(mockUser, nil)
	mockRepo.On("UpdateTokenIDs", userID, mock.Anything, mock.Anything).Return(nil)

	reqBody := map[string]string{
		"username": "mahasiswa1",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}