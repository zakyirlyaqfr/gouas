package repository

import (
	"gouas/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepository interface {
	FindAll() ([]models.Student, error)
	FindByID(id uuid.UUID) (*models.Student, error)
	FindByUserID(userID uuid.UUID) (*models.Student, error)
	UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error
	
	// [BARU] Method Tambah Poin
	AddPoints(studentID uuid.UUID, points int) error
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{db}
}

func (r *studentRepository) FindAll() ([]models.Student, error) {
	var students []models.Student
	err := r.db.Preload("User").Preload("Advisor.User").Find(&students).Error
	return students, err
}

func (r *studentRepository) FindByID(id uuid.UUID) (*models.Student, error) {
	var student models.Student
	err := r.db.Preload("User").Preload("Advisor.User").First(&student, "id = ?", id).Error
	return &student, err
}

func (r *studentRepository) FindByUserID(userID uuid.UUID) (*models.Student, error) {
	var student models.Student
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&student).Error
	return &student, err
}

func (r *studentRepository) UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	return r.db.Model(&models.Student{}).Where("id = ?", studentID).Update("advisor_id", advisorID).Error
}

// [BARU] Implementasi Tambah Poin (Atomic Update)
func (r *studentRepository) AddPoints(studentID uuid.UUID, points int) error {
	return r.db.Model(&models.Student{}).Where("id = ?", studentID).
		Update("total_points", gorm.Expr("total_points + ?", points)).Error
}