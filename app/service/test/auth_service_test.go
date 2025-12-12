package test

import (
	"gouas/app/models"
	"gouas/app/service"
	"gouas/helper"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
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

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	authService := service.NewAuthService(mockRepo)

	// Setup data dummy untuk test logic (ini memory data, bukan DB)
	password := "password123"
	hashed, _ := helper.HashPassword(password)
	
	mockUser := &models.User{
		ID:           uuid.New(),
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

	token, err := authService.Login("mahasiswa1", password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}