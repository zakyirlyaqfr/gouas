package service

import (
	"errors"
	"gouas/app/model"
	"gouas/app/repository"

	"github.com/google/uuid"
)

type AchievementService interface {
	CreateDraft(userID uuid.UUID, req model.MongoAchievement) (*model.AchievementReference, error)
	GetMyAchievements(userID uuid.UUID) ([]map[string]interface{}, error)
	SubmitAchievement(id uuid.UUID) error 
	DeleteAchievement(id uuid.UUID) error
	
	// Dosen Features
	GetAdviseeAchievements(userID uuid.UUID) ([]map[string]interface{}, error)
	VerifyAchievement(userID uuid.UUID, achievementID uuid.UUID) error
	RejectAchievement(userID uuid.UUID, achievementID uuid.UUID, note string) error
}

type achievementService struct {
	repo repository.AchievementRepository
}

func NewAchievementService(repo repository.AchievementRepository) AchievementService {
	return &achievementService{repo: repo}
}

// ... (Existing Methods: CreateDraft, GetMyAchievements, Submit, Delete) ...
// Copy-paste method lama di sini, jangan dihapus

func (s *achievementService) CreateDraft(userID uuid.UUID, req model.MongoAchievement) (*model.AchievementReference, error) {
	student, err := s.repo.FindStudentByUserID(userID)
	if err != nil { return nil, errors.New("student profile not found") }
	req.StudentID = student.ID.String()
	mongoID, err := s.repo.CreateMongo(&req)
	if err != nil { return nil, err }
	ref := &model.AchievementReference{StudentID: student.ID, MongoAchievementID: mongoID, Status: model.StatusDraft}
	if err := s.repo.CreateReference(ref); err != nil {
		s.repo.SoftDeleteMongo(mongoID)
		return nil, err
	}
	return ref, nil
}

func (s *achievementService) GetMyAchievements(userID uuid.UUID) ([]map[string]interface{}, error) {
	student, err := s.repo.FindStudentByUserID(userID)
	if err != nil { return nil, errors.New("student profile not found") }
	refs, err := s.repo.FindReferencesByStudentID(student.ID)
	if err != nil { return nil, err }
	
	var results []map[string]interface{}
	for _, ref := range refs {
		mongoData, _ := s.repo.FindMongoByID(ref.MongoAchievementID)
		item := map[string]interface{}{
			"id": ref.ID, "status": ref.Status, "mongo_id": ref.MongoAchievementID, "details": mongoData, "created_at": ref.CreatedAt,
		}
		results = append(results, item)
	}
	return results, nil
}

func (s *achievementService) SubmitAchievement(id uuid.UUID) error {
	return s.repo.UpdateReferenceStatus(id, model.StatusSubmitted)
}

func (s *achievementService) DeleteAchievement(id uuid.UUID) error {
	return s.repo.SoftDeleteReference(id)
}

// --- NEW METHODS FOR TAHAP 6 ---

func (s *achievementService) GetAdviseeAchievements(userID uuid.UUID) ([]map[string]interface{}, error) {
	// 1. Cari Profile Dosen berdasarkan User ID
	lecturer, err := s.repo.FindLecturerByUserID(userID)
	if err != nil {
		return nil, errors.New("lecturer profile not found")
	}

	// 2. Cari semua prestasi mahasiswa bimbingan beliau
	refs, err := s.repo.FindReferencesByAdvisorID(lecturer.ID)
	if err != nil {
		return nil, err
	}

	// 3. Construct Data (Include data mahasiswa)
	var results []map[string]interface{}
	for _, ref := range refs {
		mongoData, _ := s.repo.FindMongoByID(ref.MongoAchievementID)
		
		item := map[string]interface{}{
			"id":             ref.ID,
			"status":         ref.Status,
			"student_nim":    ref.Student.NIM, // Info tambahan
			"student_name":   "Loaded from Users Table", // (Simplified, harusnya join user juga)
			"details":        mongoData,
			"submitted_at":   ref.CreatedAt, // Harusnya submitted_at column, pake created_at dlu sbg mock
		}
		results = append(results, item)
	}
	return results, nil
}

func (s *achievementService) VerifyAchievement(userID uuid.UUID, achievementID uuid.UUID) error {
	// 1. Validasi Dosen
	lecturer, err := s.repo.FindLecturerByUserID(userID)
	if err != nil { return errors.New("lecturer profile not found") }

	// 2. Ambil Achievement
	ach, err := s.repo.FindReferenceByID(achievementID)
	if err != nil { return errors.New("achievement not found") }

	// 3. Validasi Hak Akses (Apakah mahasiswa ini bimbingan dosen tersebut?)
	// Pointer comparison advisorID
	if ach.Student.AdvisorID == nil || *ach.Student.AdvisorID != lecturer.ID {
		return errors.New("unauthorized: student is not your advisee")
	}

	// 4. Update Status
	return s.repo.VerifyAchievement(achievementID, userID)
}

func (s *achievementService) RejectAchievement(userID uuid.UUID, achievementID uuid.UUID, note string) error {
	// 1. Validasi Dosen
	lecturer, err := s.repo.FindLecturerByUserID(userID)
	if err != nil { return errors.New("lecturer profile not found") }

	// 2. Ambil Achievement
	ach, err := s.repo.FindReferenceByID(achievementID)
	if err != nil { return errors.New("achievement not found") }

	// 3. Validasi Hak Akses
	if ach.Student.AdvisorID == nil || *ach.Student.AdvisorID != lecturer.ID {
		return errors.New("unauthorized: student is not your advisee")
	}

	// 4. Update Status & Note
	return s.repo.RejectAchievement(achievementID, note)
}