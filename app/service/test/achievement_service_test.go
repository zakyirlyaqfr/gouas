package test

import (
	"gouas/app/models"
	"gouas/app/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCK ACHIEVEMENT REPOSITORY ---
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

func (m *MockAchievementRepo) FindAllReferences() ([]models.AchievementReference, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AchievementReference), args.Error(1)
}

func (m *MockAchievementRepo) FindReferencesByStudentID(studentID uuid.UUID) ([]models.AchievementReference, error) {
	args := m.Called(studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AchievementReference), args.Error(1)
}

// [PERBAIKAN UTAMA DI SINI]
// Ubah return type dari map[string]interface{} menjadi *models.Achievement
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

// --- MOCK STUDENT REPOSITORY ---
type MockStudentRepo struct {
	mock.Mock
}

func (m *MockStudentRepo) FindAll() ([]models.Student, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

// --- TESTS ---

func TestCreateAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	
	svc := service.NewAchievementService(mockRepo, mockStudentRepo)

	studentID := uuid.New()
	input := models.Achievement{Title: "Juara 1", AchievementType: "Competition"}
	expectedRef := &models.AchievementReference{ID: uuid.New(), Status: models.StatusDraft}

	mockRepo.On("Create", input, studentID).Return(expectedRef, nil)

	result, err := svc.Create(studentID, input)

	assert.NoError(t, err)
	assert.Equal(t, models.StatusDraft, result.Status)
	mockRepo.AssertExpectations(t)
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

	err := svc.Submit(id, studentID)
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

	// 1. Mock FindReference
	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, StudentID: studentID, Status: models.StatusSubmitted,
	}, nil)

	// 2. Mock Verify Repo
	mockRepo.On("Verify", id, verifierID).Return(nil)

	// 3. Mock AddPoints
	mockStudentRepo.On("AddPoints", studentID, 10).Return(nil)

	err := svc.Verify(id, verifierID)
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStudentRepo.AssertExpectations(t)
}

func TestRejectAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo)

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