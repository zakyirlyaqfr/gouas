package test

import (
	"bytes"
	"encoding/json"
	"gouas/app/models"
	"gouas/app/service"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAdminRepo struct {
	mock.Mock
}

func (m *MockAdminRepo) CreateUser(user models.User) (models.User, error) {
	args := m.Called(user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockAdminRepo) FindRoleByName(name string) (models.Role, error) {
	args := m.Called(name)
	return args.Get(0).(models.Role), args.Error(1)
}

// Dummy methods to satisfy interface
func (m *MockAdminRepo) UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) error { return nil }
func (m *MockAdminRepo) FindAllUsers() ([]models.User, error) { return nil, nil }
func (m *MockAdminRepo) FindUserByID(id uuid.UUID) (*models.User, error) { return nil, nil }
func (m *MockAdminRepo) UpdateUser(user models.User) error { return nil }
func (m *MockAdminRepo) DeleteUser(id uuid.UUID) error { return nil }
func (m *MockAdminRepo) CreateStudentProfile(student models.Student) error { return nil }
func (m *MockAdminRepo) CreateLecturerProfile(lecturer models.Lecturer) error { return nil }

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockAdminRepo)
	adminSvc := service.NewAdminService(mockRepo)
	app := fiber.New()
	app.Post("/users", adminSvc.CreateUser)

	roleID := uuid.New()
	mockRole := models.Role{ID: roleID, Name: "Admin"}

	mockRepo.On("FindRoleByName", "Admin").Return(mockRole, nil)
	mockRepo.On("CreateUser", mock.AnythingOfType("models.User")).Return(models.User{
		Username: "admin_baru",
		RoleID:   roleID,
	}, nil)

	reqBody := map[string]string{
		"username": "admin_baru",
		"email":    "admin@email.com",
		"password": "pass123",
		"fullName": "Admin Baru",
		"roleName": "Admin",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}