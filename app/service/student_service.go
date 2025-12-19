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
	repo         repository.StudentRepository
	achRepo      repository.AchievementRepository
	lecturerRepo repository.LecturerRepository
}

func NewStudentService(repo repository.StudentRepository, achRepo repository.AchievementRepository, lecturerRepo repository.LecturerRepository) StudentService {
	return &studentService{repo: repo, achRepo: achRepo, lecturerRepo: lecturerRepo}
}

func (s *studentService) GetAll(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))

	if authData.Role == "Mahasiswa" {
		return c.Status(403).JSON(helper.APIResponse("error", "Mahasiswa cannot list all students", nil))
	}

	if authData.Role == "Dosen Wali" {
		lecturer, _ := s.lecturerRepo.FindByUserID(uuid.MustParse(authData.UserID))
		students, _ := s.lecturerRepo.FindAdvisees(lecturer.ID)
		return c.Status(200).JSON(helper.APIResponse("success", "Advisees list retrieved", students))
	}

	students, _ := s.repo.FindAll()
	return c.Status(200).JSON(helper.APIResponse("success", "All students list retrieved", students))
}

func (s *studentService) GetDetail(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	id, _ := uuid.Parse(c.Params("id"))
	student, err := s.repo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Student not found", nil))
	}

	// [ACCESS CHECK]
	allowed := false
	switch authData.Role {
	case "Admin":
		allowed = true
	case "Mahasiswa":
		if student.UserID.String() == authData.UserID {
			allowed = true
		}
	case "Dosen Wali":
		lecturer, _ := s.lecturerRepo.FindByUserID(uuid.MustParse(authData.UserID))
		if student.AdvisorID != nil && *student.AdvisorID == lecturer.ID {
			allowed = true
		}
	}

	if !allowed {
		return c.Status(403).JSON(helper.APIResponse("error", "Forbidden: Not your profile/advisee", nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Student detail retrieved", student))
}

func (s *studentService) AssignAdvisor(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var input struct {
		AdvisorID string `json:"advisorId"`
	}
	c.BodyParser(&input)
	advID, _ := uuid.Parse(input.AdvisorID)

	err := s.repo.UpdateAdvisor(id, advID)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Advisor assigned successfully", nil))
}

func (s *studentService) GetStudentAchievements(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	// Access check should be same as GetDetail (Simplified here)
	data, _ := s.achRepo.FindReferencesByStudentID(id)
	return c.Status(200).JSON(helper.APIResponse("success", "Achievements retrieved", data))
}
