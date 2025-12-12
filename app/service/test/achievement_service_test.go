package test

import (
	"gouas/app/models"
	"gouas/app/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCK REPOSITORY ---
type MockAchievementRepo struct {
	mock.Mock
}

func (m *MockAchievementRepo) Create(data models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error) {
	args := m.Called(data, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

// --- TESTS ---

// 1. Test POST Create Achievement
func TestCreateAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	svc := service.NewAchievementService(mockRepo)

	studentID := uuid.New()
	input := models.Achievement{Title: "Juara 1", AchievementType: "Competition"}
	expectedRef := &models.AchievementReference{ID: uuid.New(), Status: models.StatusDraft}

	mockRepo.On("Create", input, studentID).Return(expectedRef, nil)

	result, err := svc.Create(studentID, input)

	assert.NoError(t, err)
	assert.Equal(t, models.StatusDraft, result.Status)
	mockRepo.AssertExpectations(t)
}

// 2. Test POST Submit (Workflow: Draft -> Submitted)
func TestSubmitAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	svc := service.NewAchievementService(mockRepo)

	id := uuid.New()
	studentID := uuid.New()

	// Mock: Find returns Draft belonging to Student
	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, StudentID: studentID, Status: models.StatusDraft,
	}, nil)

	mockRepo.On("UpdateStatus", id, models.StatusSubmitted).Return(nil)

	err := svc.Submit(id, studentID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// 3. Test POST Verify (Workflow: Submitted -> Verified)
func TestVerifyAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	svc := service.NewAchievementService(mockRepo)

	id := uuid.New()
	verifierID := uuid.New()

	// Mock: Find returns Submitted
	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, Status: models.StatusSubmitted,
	}, nil)

	mockRepo.On("Verify", id, verifierID).Return(nil)

	err := svc.Verify(id, verifierID)
	assert.NoError(t, err)
}

// 4. Test POST Reject (Workflow: Submitted -> Rejected)
func TestRejectAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	svc := service.NewAchievementService(mockRepo)

	id := uuid.New()
	verifierID := uuid.New()
	note := "Data kurang lengkap"

	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, Status: models.StatusSubmitted,
	}, nil)

	mockRepo.On("Reject", id, note).Return(nil)

	err := svc.Reject(id, verifierID, note)
	assert.NoError(t, err)
}