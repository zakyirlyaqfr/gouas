package service

import (
	"gouas/app/repository"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReportService interface {
	GetStatistics(c *fiber.Ctx) error
	GetStudentReport(c *fiber.Ctx) error
}

type reportService struct {
	repo    repository.ReportRepository
	achRepo repository.AchievementRepository
}

func NewReportService(repo repository.ReportRepository, achRepo repository.AchievementRepository) ReportService {
	return &reportService{repo: repo, achRepo: achRepo}
}

func (s *reportService) GetStatistics(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	if authData.Role == "Mahasiswa" {
		return c.Status(403).JSON(helper.APIResponse("error", "Forbidden", nil))
	}

	stats, _ := s.repo.GetAchievementStats()
	return c.Status(200).JSON(helper.APIResponse("success", "Global Statistics", stats))
}

func (s *reportService) GetStudentReport(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	// Validation owner/advisor should be here
	data, _ := s.achRepo.FindReferencesByStudentID(id)
	return c.Status(200).JSON(helper.APIResponse("success", "Student Achievement Report", data))
}