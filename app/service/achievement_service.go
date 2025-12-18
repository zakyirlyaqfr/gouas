package service

import (
	"fmt"
	"os"
	"time"

	"gouas/app/models"
	"gouas/app/repository"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AchievementService interface {
	GetAll(c *fiber.Ctx) error
	GetDetail(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	Submit(c *fiber.Ctx) error
	Verify(c *fiber.Ctx) error
	Reject(c *fiber.Ctx) error
	GetHistory(c *fiber.Ctx) error
	AddAttachment(c *fiber.Ctx) error

	// Pure Business Logic (untuk unit testing)
	CreateAchievement(data models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error)
	SubmitAchievement(id uuid.UUID, studentID uuid.UUID) error
	VerifyAchievement(id uuid.UUID, verifierID uuid.UUID) error
	RejectAchievement(id uuid.UUID, note string) error
}

type achievementService struct {
	repo        repository.AchievementRepository
	studentRepo repository.StudentRepository
}

func NewAchievementService(repo repository.AchievementRepository, studentRepo repository.StudentRepository) AchievementService {
	return &achievementService{
		repo:        repo,
		studentRepo: studentRepo,
	}
}

// ========================== PURE BUSINESS LOGIC ==========================

func (s *achievementService) CreateAchievement(data models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error) {
	if data.Title == "" || data.AchievementType == "" {
		return nil, fmt.Errorf("title and type are required")
	}
	return s.repo.Create(data, studentID)
}

func (s *achievementService) SubmitAchievement(id uuid.UUID, studentID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref.StudentID != studentID {
		return fmt.Errorf("unauthorized")
	}
	if ref.Status != models.StatusDraft {
		return fmt.Errorf("only draft can be submitted")
	}
	return s.repo.UpdateStatus(id, models.StatusSubmitted)
}

func (s *achievementService) VerifyAchievement(id uuid.UUID, verifierID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref.Status != models.StatusSubmitted {
		return fmt.Errorf("invalid achievement status")
	}
	if err := s.repo.Verify(id, verifierID); err != nil {
		return err
	}
	return s.studentRepo.AddPoints(ref.StudentID, 10)
}

func (s *achievementService) RejectAchievement(id uuid.UUID, note string) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref.Status != models.StatusSubmitted {
		return fmt.Errorf("invalid status")
	}
	return s.repo.Reject(id, note)
}

func (s *achievementService) GetAll(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	userID, _ := uuid.Parse(authData.UserID)

	var data []models.AchievementReference
	var err error

	if authData.Role == "Mahasiswa" {
		student, errProfile := s.studentRepo.FindByUserID(userID)
		if errProfile != nil {
			return c.Status(404).JSON(helper.APIResponse("error", "Student profile not found", nil))
		}
		data, err = s.repo.FindReferencesByStudentID(student.ID)
	} else {
		data, err = s.repo.FindAllReferences()
	}

	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Achievement list", data))
}

func (s *achievementService) GetDetail(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Not found", nil))
	}
	mongoData, err := s.repo.GetMongoDetail(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Detail", fiber.Map{
		"reference": ref,
		"details":   mongoData,
	}))
}

func (s *achievementService) Create(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	if !middleware.HasPermission(authData.Permissions, "achievement:create") {
		return c.Status(403).JSON(helper.APIResponse("error", "Forbidden", nil))
	}

	var input models.Achievement
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "Invalid JSON", nil))
	}

	userID, _ := uuid.Parse(authData.UserID)
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Student profile not found", nil))
	}

	result, err := s.CreateAchievement(input, student.ID)
	if err != nil {
		if err.Error() == "title and type are required" {
			return c.Status(400).JSON(helper.APIResponse("error", err.Error(), nil))
		}
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(201).JSON(helper.APIResponse("success", "Achievement created", result))
}

func (s *achievementService) Update(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	userID, _ := uuid.Parse(authData.UserID)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Profile not found", nil))
	}

	var input models.Achievement
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "Invalid JSON", nil))
	}

	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Achievement not found", nil))
	}
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(helper.APIResponse("error", "Unauthorized", nil))
	}
	if ref.Status != models.StatusDraft {
		return c.Status(400).JSON(helper.APIResponse("error", "Only draft can be updated", nil))
	}

	if err := s.repo.UpdateMongo(ref.MongoAchievementID, input); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Achievement updated", nil))
}

func (s *achievementService) Delete(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	userID, _ := uuid.Parse(authData.UserID)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Profile not found", nil))
	}

	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Achievement not found", nil))
	}
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(helper.APIResponse("error", "Unauthorized", nil))
	}
	if ref.Status != models.StatusDraft {
		return c.Status(400).JSON(helper.APIResponse("error", "Cannot delete submitted achievement", nil))
	}

	if err := s.repo.SoftDelete(id); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Achievement deleted", nil))
}

func (s *achievementService) Submit(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	userID, _ := uuid.Parse(authData.UserID)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Profile not found", nil))
	}

	id, _ := uuid.Parse(c.Params("id"))

	if err := s.SubmitAchievement(id, student.ID); err != nil {
		if err.Error() == "unauthorized" {
			return c.Status(403).JSON(helper.APIResponse("error", "Unauthorized", nil))
		}
		if err.Error() == "only draft can be submitted" {
			return c.Status(400).JSON(helper.APIResponse("error", "Only draft can be submitted", nil))
		}
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Achievement submitted", nil))
}

func (s *achievementService) Verify(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	if !middleware.HasPermission(authData.Permissions, "achievement:verify") {
		return c.Status(403).JSON(helper.APIResponse("error", "Forbidden", nil))
	}

	id, _ := uuid.Parse(c.Params("id"))
	verifierID, _ := uuid.Parse(authData.UserID)

	if err := s.VerifyAchievement(id, verifierID); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Achievement verified", nil))
}

func (s *achievementService) Reject(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	if !middleware.HasPermission(authData.Permissions, "achievement:verify") {
		return c.Status(403).JSON(helper.APIResponse("error", "Forbidden", nil))
	}

	var input struct {
		Note string `json:"note"`
	}
	if err := c.BodyParser(&input); err != nil || input.Note == "" {
		return c.Status(400).JSON(helper.APIResponse("error", "Rejection note required", nil))
	}

	id, _ := uuid.Parse(c.Params("id"))

	if err := s.RejectAchievement(id, input.Note); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Achievement rejected", nil))
}

func (s *achievementService) GetHistory(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	history := map[string]interface{}{
		"current_status":  ref.Status,
		"created_at":      ref.CreatedAt,
		"submitted_at":    ref.SubmittedAt,
		"verified_at":     ref.VerifiedAt,
		"rejected_note":   ref.RejectionNote,
	}
	return c.Status(200).JSON(helper.APIResponse("success", "History", history))
}

func (s *achievementService) AddAttachment(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	userID, _ := uuid.Parse(authData.UserID)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Profile not found", nil))
	}

	id, _ := uuid.Parse(c.Params("id"))
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil || ref.StudentID != student.ID || ref.Status != models.StatusDraft {
		return c.Status(400).JSON(helper.APIResponse("error", "Invalid achievement or status", nil))
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "File required", nil))
	}

	uniqueName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	savePath := fmt.Sprintf("./uploads/%s", uniqueName)
	fileURL := fmt.Sprintf("/uploads/%s", uniqueName)

	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", "Save failed", nil))
	}

	caption := c.FormValue("caption")
	if caption == "" {
		caption = file.Filename
	}

	attachment := models.Attachment{
		FileName:   caption,
		FileURL:    fileURL,
		FileType:   "unknown",
		UploadedAt: time.Now(),
	}

	if err := s.repo.AddAttachment(ref.MongoAchievementID, attachment); err != nil {
		os.Remove(savePath)
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Attachment added", fiber.Map{"fileUrl": fileURL}))
}