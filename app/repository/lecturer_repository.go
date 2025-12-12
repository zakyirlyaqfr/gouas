package repository

import (
	"gouas/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LecturerRepository interface {
	FindAll() ([]models.Lecturer, error)
	FindByID(id uuid.UUID) (*models.Lecturer, error)
	FindAdvisees(lecturerID uuid.UUID) ([]models.Student, error)
}

type lecturerRepository struct {
	db *gorm.DB
}

func NewLecturerRepository(db *gorm.DB) LecturerRepository {
	return &lecturerRepository{db}
}

func (r *lecturerRepository) FindAll() ([]models.Lecturer, error) {
	var lecturers []models.Lecturer
	err := r.db.Preload("User").Find(&lecturers).Error
	return lecturers, err
}

func (r *lecturerRepository) FindByID(id uuid.UUID) (*models.Lecturer, error) {
	var lecturer models.Lecturer
	err := r.db.Preload("User").First(&lecturer, "id = ?", id).Error
	return &lecturer, err
}

func (r *lecturerRepository) FindAdvisees(lecturerID uuid.UUID) ([]models.Student, error) {
	var students []models.Student
	err := r.db.Preload("User").Where("advisor_id = ?", lecturerID).Find(&students).Error
	return students, err
}