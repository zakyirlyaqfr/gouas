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
}

type achievementService struct {
	repo repository.AchievementRepository
}

func NewAchievementService(repo repository.AchievementRepository) AchievementService {
	return &achievementService{repo}
}

func (s *achievementService) Create(studentID uuid.UUID, data models.Achievement) (*models.AchievementReference, error) {
	if data.Title == "" || data.AchievementType == "" {
		return nil, errors.New("title and type are required")
	}
	return s.repo.Create(data, studentID)
}

func (s *achievementService) Submit(id uuid.UUID, studentID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	if ref.StudentID != studentID {
		return errors.New("unauthorized")
	}

	if ref.Status != models.StatusDraft {
		return errors.New("only draft achievement can be submitted")
	}

	return s.repo.UpdateStatus(id, models.StatusSubmitted)
}

func (s *achievementService) Verify(id uuid.UUID, verifierID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	if ref.Status != models.StatusSubmitted {
		return errors.New("achievement is not in submitted status")
	}

	return s.repo.Verify(id, verifierID)
}

func (s *achievementService) Reject(id uuid.UUID, verifierID uuid.UUID, note string) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	if ref.Status != models.StatusSubmitted {
		return errors.New("achievement is not in submitted status")
	}

	if note == "" {
		return errors.New("rejection note is required")
	}

	return s.repo.Reject(id, note)
}

func (s *achievementService) Delete(id uuid.UUID, studentID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	if ref.StudentID != studentID {
		return errors.New("unauthorized")
	}

	if ref.Status != models.StatusDraft {
		return errors.New("cannot delete submitted or verified achievement")
	}

	return s.repo.SoftDelete(id)
}

func (s *achievementService) AddAttachment(id uuid.UUID, studentID uuid.UUID, fileName, fileURL string) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	if ref.StudentID != studentID {
		return errors.New("unauthorized")
	}

	if ref.Status != models.StatusDraft {
		return errors.New("can only add attachments to draft")
	}

	attachment := models.Attachment{
		FileName:   fileName,
		FileURL:    fileURL,
		FileType:   "unknown",
		UploadedAt: time.Now(),
	}

	return s.repo.AddAttachment(ref.MongoAchievementID, attachment)
}