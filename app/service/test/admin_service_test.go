package test

import (
	"gouas/app/models"
	"gouas/app/service"
	"testing"

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

func (m *MockAdminRepo) UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) error {
	return nil
}

func (m *MockAdminRepo) FindAllUsers() ([]models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockAdminRepo) FindUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAdminRepo) UpdateUser(user models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockAdminRepo) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// --- NEW METHODS MOCK (Fix Error InvalidIfaceAssign) ---
func (m *MockAdminRepo) CreateStudentProfile(student models.Student) error {
	args := m.Called(student)
	return args.Error(0)
}

func (m *MockAdminRepo) CreateLecturerProfile(lecturer models.Lecturer) error {
	args := m.Called(lecturer)
	return args.Error(0)
}

// --- TESTS ---

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockAdminRepo)
	adminService := service.NewAdminService(mockRepo)

	roleID := uuid.New()
	mockRole := models.Role{ID: roleID, Name: "Admin"}

	mockRepo.On("FindRoleByName", "Admin").Return(mockRole, nil)

	// Kita match sembarang argument User karena ID & Hash bisa berubah
	mockRepo.On("CreateUser", mock.AnythingOfType("models.User")).Return(models.User{
		Username: "admin_baru",
		RoleID:   roleID,
	}, nil)

	// Note: Karena role "Admin", logic CreateStudentProfile tidak dipanggil, jadi tidak perlu di-mock expect-nya
	createdUser, err := adminService.CreateUser("admin_baru", "admin@email.com", "pass123", "Admin Baru", "Admin")

	assert.NoError(t, err)
	assert.Equal(t, "admin_baru", createdUser.Username)
	mockRepo.AssertExpectations(t)
}
