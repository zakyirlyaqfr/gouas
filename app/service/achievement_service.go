package service

import (
	"errors"
	"gouas/app/model"
	"gouas/app/repository"

	"github.com/google/uuid"
	// "go.mongodb.org/mongo-driver/bson" <-- Baris ini dihapus karena tidak terpakai
)

type AchievementService interface {
	CreateDraft(userID uuid.UUID, req model.MongoAchievement) (*model.AchievementReference, error)
	GetMyAchievements(userID uuid.UUID) ([]map[string]interface{}, error)
	SubmitAchievement(id uuid.UUID) error // ID Postgres
	DeleteAchievement(id uuid.UUID) error // ID Postgres
}

type achievementService struct {
	repo repository.AchievementRepository
}

func NewAchievementService(repo repository.AchievementRepository) AchievementService {
	return &achievementService{repo: repo}
}

func (s *achievementService) CreateDraft(userID uuid.UUID, req model.MongoAchievement) (*model.AchievementReference, error) {
	// 1. Cari Student ID berdasarkan User ID yang login
	student, err := s.repo.FindStudentByUserID(userID)
	if err != nil {
		return nil, errors.New("student profile not found for this user")
	}

	// 2. Set Student ID di Data Mongo
	req.StudentID = student.ID.String()

	// 3. Simpan ke MongoDB
	mongoID, err := s.repo.CreateMongo(&req)
	if err != nil {
		return nil, err
	}

	// 4. Simpan Referensi ke PostgreSQL (Status Draft)
	ref := &model.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: mongoID,
		Status:             model.StatusDraft,
	}

	if err := s.repo.CreateReference(ref); err != nil {
		// Idealnya: Rollback hapus data di Mongo jika Postgres gagal
		s.repo.SoftDeleteMongo(mongoID) 
		return nil, err
	}

	return ref, nil
}

func (s *achievementService) GetMyAchievements(userID uuid.UUID) ([]map[string]interface{}, error) {
	// 1. Cari Student ID
	student, err := s.repo.FindStudentByUserID(userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// 2. Ambil semua referensi dari Postgres
	refs, err := s.repo.FindReferencesByStudentID(student.ID)
	if err != nil {
		return nil, err
	}

	// 3. Gabungkan data Postgres dan Mongo
	var results []map[string]interface{}
	for _, ref := range refs {
		mongoData, _ := s.repo.FindMongoByID(ref.MongoAchievementID)
		
		// Gabungkan response
		item := map[string]interface{}{
			"id":             ref.ID, // ID Postgres untuk aksi selanjutnya
			"status":         ref.Status,
			"mongo_id":       ref.MongoAchievementID,
			"details":        mongoData, // Data lengkap dari Mongo
			"created_at":     ref.CreatedAt,
		}
		results = append(results, item)
	}

	return results, nil
}

func (s *achievementService) SubmitAchievement(id uuid.UUID) error {
	return s.repo.UpdateReferenceStatus(id, model.StatusSubmitted)
}

func (s *achievementService) DeleteAchievement(id uuid.UUID) error {
	// Di sini kita langsung update status aja sesuai instruksi simple
	return s.repo.SoftDeleteReference(id)
}