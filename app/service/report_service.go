package service

import (
	"gouas/app/repository"
	"gouas/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReportService interface {
	GetStatistics(c *fiber.Ctx) error
	GetStudentReport(c *fiber.Ctx) error
}

type reportService struct {
	repo    repository.ReportRepository
	achRepo repository.AchievementRepository // Re-use achievement repo
}

func NewReportService(repo repository.ReportRepository, achRepo repository.AchievementRepository) ReportService {
	return &reportService{repo: repo, achRepo: achRepo}
}

func (s *reportService) GetStatistics(c *fiber.Ctx) error {
	stats, err := s.repo.GetAchievementStats()
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Statistics", stats))
}

func (s *reportService) GetStudentReport(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	data, err := s.achRepo.FindReferencesByStudentID(id)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Student Report", data))
}