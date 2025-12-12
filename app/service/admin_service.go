package service

import (
	"errors"
	"fmt"
	"gouas/app/models"
	"gouas/app/repository"
	"gouas/helper"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type AdminService interface {
	CreateUser(username, email, password, fullName, roleName string) (models.User, error)
	AssignRole(userID uuid.UUID, roleName string) error
	GetAllUsers() ([]models.User, error)
	GetUserDetail(id uuid.UUID) (*models.User, error)
	UpdateUser(id uuid.UUID, fullName, email string) error
	DeleteUser(id uuid.UUID) error
}

type adminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{adminRepo}
}

// ... import "fmt" jangan lupa ditambahkan di atas

func (s *adminService) CreateUser(username, email, password, fullName, roleName string) (models.User, error) {
	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil {
		return models.User{}, errors.New("role not found")
	}

	newUser := models.User{
		Username: username, Email: email, PasswordHash: hashedPassword, FullName: fullName, RoleID: role.ID, IsActive: true,
	}

	createdUser, err := s.adminRepo.CreateUser(newUser)
	if err != nil {
		return models.User{}, err
	}

	// --- DEBUG LOG ---
	fmt.Printf("[DEBUG] User Created: %s | Role Input: '%s'\n", username, roleName)

	// --- AUTO CREATE PROFILE LOGIC ---
	randSrc := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSrc)
	randomCode := strconv.Itoa(r.Intn(90000) + 10000)

	switch roleName {
	case "Mahasiswa":
		fmt.Println("[DEBUG] Masuk logic create Profile Mahasiswa...")
		student := models.Student{
			UserID:       createdUser.ID,
			StudentID:    "NIM-" + username + "-" + randomCode,
			ProgramStudy: "Informatika",
			AcademicYear: "2025",
		}
		err := s.adminRepo.CreateStudentProfile(student)
		if err != nil {
			fmt.Printf("[ERROR] Gagal create profile: %v\n", err)
		} else {
			fmt.Println("[DEBUG] Sukses create profile Mahasiswa")
		}
	case "Dosen Wali":
		fmt.Println("[DEBUG] Masuk logic create Profile Dosen...")
		lecturer := models.Lecturer{
			UserID:     createdUser.ID,
			LecturerID: "NIP-" + username + "-" + randomCode,
			Department: "Informatika",
		}
		s.adminRepo.CreateLecturerProfile(lecturer)
	default:
		fmt.Printf("[DEBUG] Role '%s' tidak cocok dengan 'Mahasiswa' atau 'Dosen Wali', profile skip.\n", roleName)
	}

	return createdUser, nil
}

func (s *adminService) AssignRole(userID uuid.UUID, roleName string) error {
	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil {
		return errors.New("role not found")
	}
	return s.adminRepo.UpdateUserRole(userID, role.ID)
}

func (s *adminService) GetAllUsers() ([]models.User, error) {
	return s.adminRepo.FindAllUsers()
}

func (s *adminService) GetUserDetail(id uuid.UUID) (*models.User, error) {
	return s.adminRepo.FindUserByID(id)
}

func (s *adminService) UpdateUser(id uuid.UUID, fullName, email string) error {
	user, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return err
	}
	user.FullName = fullName
	user.Email = email
	return s.adminRepo.UpdateUser(*user)
}

func (s *adminService) DeleteUser(id uuid.UUID) error {
	return s.adminRepo.DeleteUser(id)
}
