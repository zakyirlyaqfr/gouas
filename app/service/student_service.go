package service

import (
	"gouas/app/repository"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentService interface {
	GetAll(c *fiber.Ctx) error
	GetDetail(c *fiber.Ctx) error
	GetStudentAchievements(c *fiber.Ctx) error
	AssignAdvisor(c *fiber.Ctx) error
}

type studentService struct {
	repo    repository.StudentRepository
	achRepo repository.AchievementRepository // Butuh akses ke achievement untuk endpoint get achievement
}

// Perlu inject Achievement Repo juga jika endpoint get achievement ada di student route
func NewStudentService(repo repository.StudentRepository, achRepo repository.AchievementRepository) StudentService {
	return &studentService{repo: repo, achRepo: achRepo}
}

func (s *studentService) GetAll(c *fiber.Ctx) error {
	students, err := s.repo.FindAll()
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Student list", students))
}

func (s *studentService) GetDetail(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	student, err := s.repo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Not found", nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Student detail", student))
}

func (s *studentService) AssignAdvisor(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	if authData.Role != "Admin" {
		return c.Status(403).JSON(helper.APIResponse("error", "Forbidden", nil))
	}

	id, _ := uuid.Parse(c.Params("id"))
	var input struct {
		AdvisorID string `json:"advisorId"`
	}
	c.BodyParser(&input)
	advID, _ := uuid.Parse(input.AdvisorID)

	if err := s.repo.UpdateAdvisor(id, advID); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Advisor assigned", nil))
}

func (s *studentService) GetStudentAchievements(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	data, err := s.achRepo.FindReferencesByStudentID(id)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Student achievements", data))
}