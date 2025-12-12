package service

import (
	"gouas/app/models"
	"gouas/app/repository"

	"github.com/google/uuid"
)

type LecturerService interface {
	GetAll() ([]models.Lecturer, error)
	GetAdvisees(lecturerID uuid.UUID) ([]models.Student, error)
}

type lecturerService struct {
	repo repository.LecturerRepository
}

func NewLecturerService(repo repository.LecturerRepository) LecturerService {
	return &lecturerService{repo}
}

func (s *lecturerService) GetAll() ([]models.Lecturer, error) {
	return s.repo.FindAll()
}

func (s *lecturerService) GetAdvisees(lecturerID uuid.UUID) ([]models.Student, error) {
	return s.repo.FindAdvisees(lecturerID)
}