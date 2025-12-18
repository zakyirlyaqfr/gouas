package service

import (
	"gouas/app/models"
	"gouas/app/repository"
	"gouas/helper"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminService interface {
	CreateUser(c *fiber.Ctx) error
	AssignRole(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
	GetUserDetail(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

type adminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{adminRepo}
}

func (s *adminService) CreateUser(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
		RoleName string `json:"roleName"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "Invalid input", nil))
	}

	hashedPassword, err := helper.HashPassword(input.Password)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", "Hashing failed", nil))
	}

	role, err := s.adminRepo.FindRoleByName(input.RoleName)
	if err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "Role not found", nil))
	}

	newUser := models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		FullName:     input.FullName,
		RoleID:       role.ID,
		IsActive:     true,
	}

	createdUser, err := s.adminRepo.CreateUser(newUser)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	// AUTO-CREATE PROFILE
	randSrc := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSrc)
	randomCode := strconv.Itoa(r.Intn(90000) + 10000)

	switch input.RoleName {
	case "Mahasiswa":
		student := models.Student{
			UserID:       createdUser.ID,
			NIM:          "NIM-" + input.Username + "-" + randomCode,
			ProgramStudy: "Informatika",
			AcademicYear: "2025",
		}
		if err := s.adminRepo.CreateStudentProfile(student); err != nil {
			s.adminRepo.DeleteUser(createdUser.ID)
			return c.Status(500).JSON(helper.APIResponse("error", "Failed to create student profile", nil))
		}
	case "Dosen Wali":
		lecturer := models.Lecturer{
			UserID:     createdUser.ID,
			NIP:        "NIP-" + input.Username + "-" + randomCode,
			Department: "Informatika",
		}
		if err := s.adminRepo.CreateLecturerProfile(lecturer); err != nil {
			s.adminRepo.DeleteUser(createdUser.ID)
			return c.Status(500).JSON(helper.APIResponse("error", "Failed to create lecturer profile", nil))
		}
	}

	return c.Status(201).JSON(helper.APIResponse("success", "User created", createdUser))
}

func (s *adminService) AssignRole(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var input struct {
		RoleName string `json:"roleName"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "Invalid input", nil))
	}

	role, err := s.adminRepo.FindRoleByName(input.RoleName)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Role not found", nil))
	}
	if err := s.adminRepo.UpdateUserRole(id, role.ID); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Role updated", nil))
}

func (s *adminService) GetAllUsers(c *fiber.Ctx) error {
	users, err := s.adminRepo.FindAllUsers()
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "User list", users))
}

func (s *adminService) GetUserDetail(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	user, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "User not found", nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "User detail", user))
}

func (s *adminService) UpdateUser(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var input struct {
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	}
	c.BodyParser(&input)

	user, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "User not found", nil))
	}
	user.FullName = input.FullName
	user.Email = input.Email

	if err := s.adminRepo.UpdateUser(*user); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "User updated", nil))
}

func (s *adminService) DeleteUser(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	if err := s.adminRepo.DeleteUser(id); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "User deleted", nil))
}