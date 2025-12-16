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

func (m *MockAuthRepo) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// [PERBAIKAN 1] Tambahkan Mock untuk UpdateTokenIDs agar sesuai interface baru
func (m *MockAuthRepo) UpdateTokenIDs(userID uuid.UUID, accessID *uuid.UUID, refreshID *uuid.UUID) error {
	// Kita gunakan mock.Anything di test case nanti untuk argument pointer
	args := m.Called(userID, accessID, refreshID)
	return args.Error(0)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	authService := service.NewAuthService(mockRepo)

	// Setup data dummy
	password := "password123"
	hashed, _ := helper.HashPassword(password)
	userID := uuid.New() // Simpan ID agar konsisten

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

	// Expectation 1: Cari User
	mockRepo.On("FindByUsername", "mahasiswa1").Return(mockUser, nil)

	// [PERBAIKAN 2] Expectation 2: Update Token IDs harus dipanggil saat Login
	// Kita gunakan mock.Anything untuk parameter AccessID dan RefreshID karena itu digenerate random di service
	mockRepo.On("UpdateTokenIDs", userID, mock.Anything, mock.Anything).Return(nil)

	// Action
	accessToken, refreshToken, err := authService.Login("mahasiswa1", password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken, "Access Token tidak boleh kosong")
	assert.NotEmpty(t, refreshToken, "Refresh Token tidak boleh kosong")
	
	mockRepo.AssertExpectations(t)
}