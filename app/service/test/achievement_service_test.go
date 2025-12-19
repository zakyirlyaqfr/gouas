package test

import (
	"testing"

	"gouas/app/models"
	"gouas/app/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- MOCK ACHIEVEMENT REPOSITORY ---
type MockAchievementRepo struct {
	mock.Mock
}

func (m *MockAchievementRepo) Create(data models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error) {
	args := m.Called(data, studentID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.AchievementReference), args.Error(1)
}
func (m *MockAchievementRepo) FindReferenceByID(id uuid.UUID) (*models.AchievementReference, error) {
	args := m.Called(id)
	if args.Get(0) == nil { return nil, args.Error(1) }
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
func (m *MockAchievementRepo) FindReferencesByStudentIDs(studentIDs []uuid.UUID) ([]models.AchievementReference, error) {
	args := m.Called(studentIDs)
	return args.Get(0).([]models.AchievementReference), args.Error(1)
}
func (m *MockAchievementRepo) GetMongoDetail(mongoID string) (*models.Achievement, error) {
	args := m.Called(mongoID)
	if args.Get(0) == nil { return nil, args.Error(1) }
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
	return args.Get(0).([]models.Student), args.Error(1)
}
func (m *MockStudentRepo) FindByID(id uuid.UUID) (*models.Student, error) {
	args := m.Called(id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Student), args.Error(1)
}
func (m *MockStudentRepo) FindByUserID(userID uuid.UUID) (*models.Student, error) {
	args := m.Called(userID)
	if args.Get(0) == nil { return nil, args.Error(1) }
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

// --- [BARU] MOCK LECTURER REPOSITORY ---
type MockLecturerRepo struct {
	mock.Mock
}

func (m *MockLecturerRepo) FindAll() ([]models.Lecturer, error) {
	args := m.Called()
	return args.Get(0).([]models.Lecturer), args.Error(1)
}
func (m *MockLecturerRepo) FindByID(id uuid.UUID) (*models.Lecturer, error) {
	args := m.Called(id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Lecturer), args.Error(1)
}
func (m *MockLecturerRepo) FindByUserID(userID uuid.UUID) (*models.Lecturer, error) {
	args := m.Called(userID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Lecturer), args.Error(1)
}
func (m *MockLecturerRepo) FindAdvisees(lecturerID uuid.UUID) ([]models.Student, error) {
	args := m.Called(lecturerID)
	return args.Get(0).([]models.Student), args.Error(1)
}

// ==================== TESTS ====================

func TestCreateAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo, mockLecturerRepo)

	studentID := uuid.New()
	achievementData := models.Achievement{
		Title:           "Juara Lomba Pemrograman",
		AchievementType: "Kompetisi",
		Points:          10, // [FIX] Tambahkan poin agar lolos validasi
		Tags:            []string{"coding"},
	}

	expectedRef := &models.AchievementReference{
		ID:        uuid.New(),
		StudentID: studentID,
		Status:    models.StatusDraft,
	}

	// Pastikan mock On mencocokkan data yang sama (termasuk Points)
	mockRepo.On("Create", achievementData, studentID).Return(expectedRef, nil)

	result, err := svc.CreateAchievement(achievementData, studentID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedRef, result)
	mockRepo.AssertExpectations(t)
}

func TestVerifyAchievement_Success(t *testing.T) {
	// 1. Inisialisasi Mock
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo, mockLecturerRepo)

	// 2. Setup Variable Dummy
	id := uuid.New()
	verifierUserID := uuid.New()   // ID User Dosen yang login
	studentID := uuid.New()        // ID Profile Student
	lecturerProfileID := uuid.New() // ID Profile Dosen
	mongoID := "657f1a2b3c4d5e6f7a8b9c0d" // Contoh ID Mongo
	
	// Kita test skenario "National" yang bernilai 30 Poin
	expectedLevel := "National"
	expectedPoints := 30 

	// 3. Mock Setup: FindReferenceByID (Postgres)
	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID:                 id,
		StudentID:          studentID,
		MongoAchievementID: mongoID,
		Status:             models.StatusSubmitted,
	}, nil)

	// 4. Mock Setup: Validasi Advisor (Postgres)
	// Mock Cari data mahasiswa untuk cek siapa dosen walinya
	mockStudentRepo.On("FindByID", studentID).Return(&models.Student{
		ID:        studentID,
		AdvisorID: &lecturerProfileID,
	}, nil)
	
	// Mock Cari profil dosen yang sedang login
	mockLecturerRepo.On("FindByUserID", verifierUserID).Return(&models.Lecturer{
		ID: lecturerProfileID,
	}, nil)

	// 5. Mock Setup: Get Detail dari MongoDB untuk ambil LEVEL
	mockRepo.On("GetMongoDetail", mongoID).Return(&models.Achievement{
		ID:    primitive.NewObjectID(), // Jika menggunakan primitive, pastikan import 
		Title: "Juara Lomba Nasional",
		Details: models.AchievementDetails{
			CompetitionLevel: expectedLevel,
		},
	}, nil)

	// 6. Mock Setup: Eksekusi Update Status & Tambah Poin
	mockRepo.On("Verify", id, verifierUserID).Return(nil)
	
	// Pastikan poin yang dipanggil adalah 30 (Sesuai level National)
	mockStudentRepo.On("AddPoints", studentID, expectedPoints).Return(nil)

	// 7. Eksekusi Fungsi yang di-test
	err := svc.VerifyAchievement(id, verifierUserID)

	// 8. Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStudentRepo.AssertExpectations(t)
	mockLecturerRepo.AssertExpectations(t)
}

func TestRejectAchievement_Success(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo, mockLecturerRepo)

	id := uuid.New()
	verifierUserID := uuid.New()
	studentID := uuid.New()
	lecturerProfileID := uuid.New()
	note := "Data kurang lengkap"

	mockRepo.On("FindReferenceByID", id).Return(&models.AchievementReference{
		ID: id, StudentID: studentID, Status: models.StatusSubmitted,
	}, nil)

	mockStudentRepo.On("FindByID", studentID).Return(&models.Student{
		ID: studentID, AdvisorID: &lecturerProfileID,
	}, nil)

	mockLecturerRepo.On("FindByUserID", verifierUserID).Return(&models.Lecturer{
		ID: lecturerProfileID,
	}, nil)

	mockRepo.On("Reject", id, note).Return(nil)

	err := svc.RejectAchievement(id, verifierUserID, note) // Note: VerifierID di logic reject murni biasanya diproses di handler/bridge

	assert.NoError(t, err)
}

func TestCreateAchievement_PointsValidationError(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo, mockLecturerRepo)

	studentID := uuid.New()
	// Input tanpa points (points = 0)
	invalidData := models.Achievement{
		Title:           "Juara 1",
		AchievementType: "Competition",
		Points:          0, 
	}

	result, err := svc.CreateAchievement(invalidData, studentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "points field is required")
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateAchievement_SuccessWithPoints(t *testing.T) {
	mockRepo := new(MockAchievementRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	svc := service.NewAchievementService(mockRepo, mockStudentRepo, mockLecturerRepo)

	studentID := uuid.New()
	achievementData := models.Achievement{
		Title:           "Juara 1",
		AchievementType: "Competition",
		Points:          10, // Mengisi poin
		Tags:            []string{"coding"},
	}

	expectedRef := &models.AchievementReference{ID: uuid.New(), Status: models.StatusDraft}
	mockRepo.On("Create", achievementData, studentID).Return(expectedRef, nil)

	result, err := svc.CreateAchievement(achievementData, studentID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}