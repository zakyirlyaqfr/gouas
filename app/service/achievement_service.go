package service

import (
	"errors"
	"gouas/app/models"
	"gouas/app/repository"
	"time"

	"github.com/google/uuid"
)

type AchievementService interface {
	Create(studentID uuid.UUID, data models.Achievement) (*models.AchievementReference, error)
	Submit(id uuid.UUID, studentID uuid.UUID) error
	Verify(id uuid.UUID, verifierID uuid.UUID) error
	Reject(id uuid.UUID, verifierID uuid.UUID, note string) error
	Delete(id uuid.UUID, studentID uuid.UUID) error
	AddAttachment(id uuid.UUID, studentID uuid.UUID, fileName, fileURL string) error
	// New
	GetAll(role string, userID uuid.UUID) ([]models.AchievementReference, error)
	GetDetail(id uuid.UUID) (map[string]interface{}, error)
	Update(id uuid.UUID, studentID uuid.UUID, data models.Achievement) error
	GetHistory(id uuid.UUID) (map[string]interface{}, error)
}

type achievementService struct {
	repo        repository.AchievementRepository
	studentRepo repository.StudentRepository // [BARU] Inject StudentRepo
}

// [UBAH] Constructor menerima StudentRepository juga
func NewAchievementService(repo repository.AchievementRepository, studentRepo repository.StudentRepository) AchievementService {
	return &achievementService{
		repo:        repo,
		studentRepo: studentRepo,
	}
}

func (s *achievementService) Create(studentID uuid.UUID, data models.Achievement) (*models.AchievementReference, error) {
	if data.Title == "" || data.AchievementType == "" {
		return nil, errors.New("title and type are required")
	}
	return s.repo.Create(data, studentID)
}

func (s *achievementService) Submit(id uuid.UUID, studentID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return err }
	if ref.StudentID != studentID { return errors.New("unauthorized") }
	if ref.Status != models.StatusDraft { return errors.New("only draft achievement can be submitted") }
	return s.repo.UpdateStatus(id, models.StatusSubmitted)
}

// [UBAH] Logic Verify menambahkan Poin
func (s *achievementService) Verify(id uuid.UUID, verifierID uuid.UUID) error {
	// 1. Ambil data untuk dapat StudentID
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return err }
	
	if ref.Status != models.StatusSubmitted { 
		return errors.New("achievement is not in submitted status") 
	}

	// 2. Update Status jadi Verified
	if err := s.repo.Verify(id, verifierID); err != nil {
		return err
	}

	// 3. [BARU] Tambah Poin (Misal flat 10 poin)
	points := 10
	return s.studentRepo.AddPoints(ref.StudentID, points)
}

func (s *achievementService) Reject(id uuid.UUID, verifierID uuid.UUID, note string) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return err }
	if ref.Status != models.StatusSubmitted { return errors.New("achievement is not in submitted status") }
	if note == "" { return errors.New("rejection note is required") }
	return s.repo.Reject(id, note)
}

func (s *achievementService) Delete(id uuid.UUID, studentID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return err }
	if ref.StudentID != studentID { return errors.New("unauthorized") }
	if ref.Status != models.StatusDraft { return errors.New("cannot delete submitted or verified achievement") }
	return s.repo.SoftDelete(id)
}

func (s *achievementService) AddAttachment(id uuid.UUID, studentID uuid.UUID, fileName, fileURL string) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return err }
	if ref.StudentID != studentID { return errors.New("unauthorized") }
	if ref.Status != models.StatusDraft { return errors.New("can only add attachments to draft") }
	attachment := models.Attachment{FileName: fileName, FileURL: fileURL, FileType: "unknown", UploadedAt: time.Now()}
	return s.repo.AddAttachment(ref.MongoAchievementID, attachment)
}

// --- NEW IMPL ---

func (s *achievementService) GetAll(role string, userID uuid.UUID) ([]models.AchievementReference, error) {
	// Jika mahasiswa, hanya lihat punya sendiri
	if role == "Mahasiswa" {
		return s.repo.FindReferencesByStudentID(userID)
	}
	return s.repo.FindAllReferences()
}

func (s *achievementService) GetDetail(id uuid.UUID) (map[string]interface{}, error) {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return nil, err }

	mongoData, err := s.repo.GetMongoDetail(ref.MongoAchievementID)
	if err != nil { return nil, err }

	return map[string]interface{}{
		"reference": ref,
		"details":   mongoData,
	}, nil
}

func (s *achievementService) Update(id uuid.UUID, studentID uuid.UUID, data models.Achievement) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return err }
	if ref.StudentID != studentID { return errors.New("unauthorized") }
	if ref.Status != models.StatusDraft { return errors.New("only draft can be updated") }

	return s.repo.UpdateMongo(ref.MongoAchievementID, data)
}

func (s *achievementService) GetHistory(id uuid.UUID) (map[string]interface{}, error) {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil { return nil, err }

	return map[string]interface{}{
		"current_status": ref.Status,
		"created_at":     ref.CreatedAt,
		"submitted_at":   ref.SubmittedAt,
		"verified_at":    ref.VerifiedAt,
		"rejected_note":  ref.RejectionNote,
	}, nil
}