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

// === LOGIC UTAMA DISINI ===
func (s *adminService) CreateUser(username, email, password, fullName, roleName string) (models.User, error) {
	// 1. Hash Password
	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	// 2. Cari Role ID
	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil {
		return models.User{}, errors.New("role not found (pastikan 'Mahasiswa' atau 'Dosen Wali')")
	}

	// 3. Create User di Tabel Users
	newUser := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		RoleID:       role.ID,
		IsActive:     true,
	}

	createdUser, err := s.adminRepo.CreateUser(newUser)
	if err != nil {
		return models.User{}, err
	}

	fmt.Printf("[DEBUG] User Created: %s (ID: %s) | Role: %s\n", username, createdUser.ID, roleName)

	// 4. AUTO-CREATE PROFILE (Logic Otomatis)
	// Kita generate angka random untuk NIM/NIP sementara
	randSrc := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSrc)
	randomCode := strconv.Itoa(r.Intn(90000) + 10000) // Contoh: 48291

	switch roleName {
	case "Mahasiswa":
		// Buat Struct Student
		student := models.Student{
			UserID:       createdUser.ID,                       // Link ke User yang baru dibuat
			NIM:    "NIM-" + username + "-" + randomCode, // Generate NIM String
			ProgramStudy: "Informatika",                        // Default sementara
			AcademicYear: "2025",
		}
		
		// Simpan ke Tabel Students
		if err := s.adminRepo.CreateStudentProfile(student); err != nil {
			// Jika gagal buat profile, idealnya user juga dihapus (rollback manual)
			fmt.Printf("[ERROR] Gagal membuat profile Mahasiswa: %v\n", err)
			s.adminRepo.DeleteUser(createdUser.ID) 
			return models.User{}, errors.New("failed to create student profile, rolling back user")
		}
		fmt.Println("[DEBUG] Sukses create profile Mahasiswa di tabel students")

	case "Dosen Wali":
		// Buat Struct Lecturer
		lecturer := models.Lecturer{
			UserID:     createdUser.ID,
			NIP: "NIP-" + username + "-" + randomCode,
			Department: "Informatika",
		}
		
		// Simpan ke Tabel Lecturers
		if err := s.adminRepo.CreateLecturerProfile(lecturer); err != nil {
			fmt.Printf("[ERROR] Gagal membuat profile Dosen: %v\n", err)
			s.adminRepo.DeleteUser(createdUser.ID)
			return models.User{}, errors.New("failed to create lecturer profile, rolling back user")
		}
		fmt.Println("[DEBUG] Sukses create profile Dosen di tabel lecturers")
	}

	return createdUser, nil
}

// ... Fungsi sisanya (AssignRole, GetAllUsers, dll) biarkan sama ...
func (s *adminService) AssignRole(userID uuid.UUID, roleName string) error {
	role, err := s.adminRepo.FindRoleByName(roleName)
	if err != nil { return errors.New("role not found") }
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
	if err != nil { return err }
	user.FullName = fullName
	user.Email = email
	return s.adminRepo.UpdateUser(*user)
}

func (s *adminService) DeleteUser(id uuid.UUID) error {
	return s.adminRepo.DeleteUser(id)
}