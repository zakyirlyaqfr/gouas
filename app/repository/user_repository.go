package repository

import (
	"gouas/app/model"
	"gouas/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	// User Basic
	GetAllUsers() ([]model.User, error)
	FindUserByID(id uuid.UUID) (*model.User, error)
	UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) error
	DeleteUser(userID uuid.UUID) error
	
	// Profiles
	CreateOrUpdateStudent(student *model.Student) error
	CreateOrUpdateLecturer(lecturer *model.Lecturer) error
	
	// Advisor
	AssignAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: database.DB}
}

func (r *userRepository) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

func (r *userRepository) FindUserByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").First(&user, "id = ?", id).Error
	return &user, err
}

func (r *userRepository) UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}

func (r *userRepository) DeleteUser(userID uuid.UUID) error {
	return r.db.Delete(&model.User{}, "id = ?", userID).Error
}

func (r *userRepository) CreateOrUpdateStudent(student *model.Student) error {
	// Upsert: Jika ada update, jika tidak create
	return r.db.Where(model.Student{UserID: student.UserID}).
		Assign(model.Student{
			NIM:          student.NIM, // <--- Perhatikan perubahan ini (dulunya StudentID)
			ProgramStudy: student.ProgramStudy,
			AcademicYear: student.AcademicYear,
		}).
		FirstOrCreate(student).Error
}

func (r *userRepository) CreateOrUpdateLecturer(lecturer *model.Lecturer) error {
	return r.db.Where(model.Lecturer{UserID: lecturer.UserID}).
		Assign(model.Lecturer{
			LecturerID: lecturer.LecturerID,
			Department: lecturer.Department,
		}).
		FirstOrCreate(lecturer).Error
}

func (r *userRepository) AssignAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	// Update advisor_id di tabel students
	return r.db.Model(&model.Student{}).Where("id = ?", studentID).Update("advisor_id", advisorID).Error
}