package service

import (
	"gouas/app/models"
	"gouas/app/repository"

	"github.com/google/uuid"
)

type StudentService interface {
	GetAll() ([]models.Student, error)
	GetDetail(id uuid.UUID) (*models.Student, error)
	AssignAdvisor(studentID, advisorID uuid.UUID) error
	// NEW
	GetProfileByUserID(userID uuid.UUID) (*models.Student, error)
}

type studentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) StudentService {
	return &studentService{repo}
}

func (s *studentService) GetAll() ([]models.Student, error) {
	return s.repo.FindAll()
}

func (s *studentService) GetDetail(id uuid.UUID) (*models.Student, error) {
	return s.repo.FindByID(id)
}

func (s *studentService) AssignAdvisor(studentID, advisorID uuid.UUID) error {
	return s.repo.UpdateAdvisor(studentID, advisorID)
}

// --- NEW IMPL ---
func (s *studentService) GetProfileByUserID(userID uuid.UUID) (*models.Student, error) {
	return s.repo.FindByUserID(userID)
}
