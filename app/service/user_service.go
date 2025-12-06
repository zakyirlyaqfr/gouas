package service

import (
	"gouas/app/model"
	"gouas/app/repository"
	"gouas/utils"

	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(username, email, password, fullName string, roleID uuid.UUID) (*model.User, error)
	GetAllUsers() ([]model.User, error)
	ChangeRole(userID uuid.UUID, roleID uuid.UUID) error
	SetupStudentProfile(userID uuid.UUID, nim, studyProgram, year string) error
	SetupLecturerProfile(userID uuid.UUID, nip, department string) error
	AssignAdvisor(studentID uuid.UUID, lecturerID uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository // Reuse untuk create user
}

func NewUserService(userRepo repository.UserRepository, authRepo repository.AuthRepository) UserService {
	return &userService{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

func (s *userService) CreateUser(username, email, password, fullName string, roleID uuid.UUID) (*model.User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		RoleID:       roleID,
		IsActive:     true,
	}

	if err := s.authRepo.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetAllUsers() ([]model.User, error) {
	return s.userRepo.GetAllUsers()
}

func (s *userService) ChangeRole(userID uuid.UUID, roleID uuid.UUID) error {
	return s.userRepo.UpdateUserRole(userID, roleID)
}

func (s *userService) SetupStudentProfile(userID uuid.UUID, nim, studyProgram, year string) error {
	student := &model.Student{
		UserID:       userID,
		NIM:          nim, // <--- Perhatikan perubahan ini
		ProgramStudy: studyProgram,
		AcademicYear: year,
	}
	return s.userRepo.CreateOrUpdateStudent(student)
}

func (s *userService) SetupLecturerProfile(userID uuid.UUID, nip, department string) error {
	lecturer := &model.Lecturer{
		UserID:     userID,
		LecturerID: nip,
		Department: department,
	}
	return s.userRepo.CreateOrUpdateLecturer(lecturer)
}

func (s *userService) AssignAdvisor(studentID uuid.UUID, lecturerID uuid.UUID) error {
	// Di sini studentID adalah ID dari tabel students (bukan user_id)
	// lecturerID adalah ID dari tabel lecturers (bukan user_id)
	return s.userRepo.AssignAdvisor(studentID, lecturerID)
}