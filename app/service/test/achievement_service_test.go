package test

import (
	"testing"

	"gouas/app/models"
	"gouas/app/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Achievement Repository (pastikan semua method ada)
type MockAchievementRepo struct {
	mock.Mock
}

func (m *MockAchievementRepo) Create(data models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error) {
	args := m.Called(data, studentID)
	return args.Get(0).(*models.AchievementReference), args.Error(1)
}
func (m *MockAchievementRepo) FindReferenceByID(id uuid.UUID) (*models.AchievementReference, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AchievementReference), args.Error(1)
}
func (m *MockAchievementRepo) UpdateStatus(id uuid.UUID, status models.AchievementStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}
func (m *MockAchievementRepo) Verify(id uuid.UUID, verifierID uuid.UUID) error {
	args := m.Called(id, verifierID)
	return args.Error(0)
}
func (m *MockAchievementRepo) Reject(id uuid.UUID, note string) error {
	args := m.Called(id, note)
	return args.Error(0)
}
func (m *MockAchievementRepo) AddAttachment(mongoID string, attachment models.Attachment) error {
	args := m.Called(mongoID, attachment)
	return args.Error(0)
}
func (m *MockAchievementRepo) SoftDelete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockAchievementRepo) FindAllReferences() ([]models.AchievementReference, error) {
	args := m.Called()
	return args.Get(0).([]models.AchievementReference), args.Error(1)
}
func (m *MockAchievementRepo) FindReferencesByStudentID(studentID uuid.UUID) ([]models.AchievementReference, error) {
	args := m.Called(studentID)
	return args.Get(0).([]models.AchievementReference), args.Error(1)
}
func (m *MockAchievementRepo) GetMongoDetail(mongoID string) (*models.Achievement, error) {
	args := m.Called(mongoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Achievement), args.Error(1)
}
func (m *MockAchievementRepo) UpdateMongo(mongoID string, data models.Achievement) error {
	args := m.Called(mongoID, data)
	return args.Error(0)
}

// Mock Student Repository
type MockStudentRepo struct {
	mock.Mock
}

func (m *MockStudentRepo) FindAll() ([]models.Student, error) {
	args := m.Called()
	return args.Get(0).([]models.Student), args.Error(1)
}
func (m *MockStudentRepo) FindByID(id uuid.UUID) (*models.Student, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Student), args.Error(1)
}
func (m *MockStudentRepo) FindByUserID(userID uuid.UUID) (*models.Student, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Student), args.Error(1)
}
func (m *MockStudentRepo) UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	args := m.Called(studentID, advisorID)
	return args.Error(0)
}
func (m *MockStudentRepo) AddPoints(studentID uuid.UUID, points int) error {
	args := m.Called(studentID, points)
	return args.Error(0)
}

// ==================== TESTS ====================

func TestCreateAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo)

	studentID := uuid.New()
	achievementData := models.Achievement{
		Title:           "Juara Lomba Pemrograman",
		AchievementType: "Kompetisi",
	}

	expectedRef := &models.AchievementReference{
		ID:      uuid.New(),
		StudentID: studentID,
		Status:  models.StatusDraft,
	}

	mockRepo.On("Create", achievementData, studentID).Return(expectedRef, nil)

	result, err := svc.CreateAchievement(achievementData, studentID)

	assert.NoError(t, err)
	assert.Equal(t, expectedRef, result)
	mockRepo.AssertExpectations(t)
}

func TestCreateAchievement_ValidationError(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo)

	studentID := uuid.New()
	invalidData := models.Achievement{
		Title: "", // kosong
	}

	_, err := svc.CreateAchievement(invalidData, studentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title and type are required")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestSubmitAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo)

	id := uuid.New()
	studentID := uuid.New()

	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, StudentID: studentID, Status: models.StatusDraft,
	}, nil)

	mockRepo.On("UpdateStatus", id, models.StatusSubmitted).Return(nil)

	err := svc.SubmitAchievement(id, studentID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestVerifyAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo)

	id := uuid.New()
	verifierID := uuid.New()
	studentID := uuid.New()

	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, StudentID: studentID, Status: models.StatusSubmitted,
	}, nil)

	mockRepo.On("Verify", id, verifierID).Return(nil)
	mockStudentRepo.On("AddPoints", studentID, 10).Return(nil)

	err := svc.VerifyAchievement(id, verifierID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStudentRepo.AssertExpectations(t)
}

func TestRejectAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo)

	id := uuid.New()
	note := "Data kurang lengkap"

	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, Status: models.StatusSubmitted,
	}, nil)

	mockRepo.On("Reject", id, note).Return(nil)

	err := svc.RejectAchievement(id, note)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}