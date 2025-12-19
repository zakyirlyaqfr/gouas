package service

import (
	"fmt"
	"time"

	"gouas/app/models"
	"gouas/app/repository"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AchievementService interface {
	// Handler methods (Menerima Fiber Ctx)
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

	// Pure Business Logic (Murni Logic, untuk Unit Test)
	CreateAchievement(data models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error)
	SubmitAchievement(id uuid.UUID, studentID uuid.UUID) error
	VerifyAchievement(id uuid.UUID, verifierUserID uuid.UUID) error
	RejectAchievement(id uuid.UUID, verifierUserID uuid.UUID, note string) error
}

type achievementService struct {
	repo         repository.AchievementRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewAchievementService(repo repository.AchievementRepository, studentRepo repository.StudentRepository, lecturerRepo repository.LecturerRepository) AchievementService {
	return &achievementService{
		repo:         repo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// =========================================================================
// 1. PURE BUSINESS LOGIC (Untuk di-test di Unit Test)
// =========================================================================

func (s *achievementService) CreateAchievement(data models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error) {
	// Validasi Title & Type
	if data.Title == "" || data.AchievementType == "" {
		return nil, fmt.Errorf("title and type are required")
	}

	// [BARU] Validasi Points Wajib diisi dan harus positif
	// Karena tipe int, jika tidak dikirim akan bernilai 0.
	if data.Points <= 0 {
		return nil, fmt.Errorf("points field is required and must be a positive number (minimum 1)")
	}

	return s.repo.Create(data, studentID)
}

func (s *achievementService) SubmitAchievement(id uuid.UUID, studentID uuid.UUID) error {
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return fmt.Errorf("achievement not found")
	}
	if ref.StudentID != studentID {
		return fmt.Errorf("unauthorized: you don't own this")
	}
	if ref.Status != models.StatusDraft && ref.Status != models.StatusRejected {
		return fmt.Errorf("only draft or rejected can be submitted")
	}
	return s.repo.UpdateStatus(id, models.StatusSubmitted)
}

func (s *achievementService) VerifyAchievement(id uuid.UUID, verifierUserID uuid.UUID) error {
	// 1. Ambil Reference dari Postgres
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil || ref.Status != models.StatusSubmitted {
		return fmt.Errorf("invalid achievement or status")
	}

	// 2. Validasi Dosen Wali (Logic tetap sama)
	student, _ := s.studentRepo.FindByID(ref.StudentID)
	lecturer, errL := s.lecturerRepo.FindByUserID(verifierUserID)
	if errL != nil {
		return fmt.Errorf("lecturer profile not found")
	}
	if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
		return fmt.Errorf("forbidden: you are not the advisor for this student")
	}

	// 3. [BARU] Ambil Detail dari MongoDB untuk Cek Level
	mongoDetail, errM := s.repo.GetMongoDetail(ref.MongoAchievementID)
	if errM != nil {
		return fmt.Errorf("could not fetch achievement details from mongo")
	}

	// 4. [BARU] Tentukan Poin berdasarkan CompetitionLevel
	var pointAwarded int
	switch mongoDetail.Details.CompetitionLevel {
	case "International":
		pointAwarded = 50
	case "National":
		pointAwarded = 30
	case "Provincial":
		pointAwarded = 20
	case "Campus":
		pointAwarded = 10
	default:
		pointAwarded = 5 // Poin dasar jika level tidak diisi/lainnya
	}

	// 5. Update Status di Postgres
	if err := s.repo.Verify(id, verifierUserID); err != nil {
		return err
	}
	
	// 6. [BARU] Tambah Poin sesuai Level yang sudah dihitung
	return s.studentRepo.AddPoints(ref.StudentID, pointAwarded)
}

func (s *achievementService) RejectAchievement(id uuid.UUID, verifierUserID uuid.UUID, note string) error {
	if note == "" {
		return fmt.Errorf("rejection note is required")
	}

	ref, err := s.repo.FindReferenceByID(id)
	if err != nil || ref.Status != models.StatusSubmitted {
		return fmt.Errorf("invalid achievement or status")
	}

	// Validasi Dosen Wali
	student, _ := s.studentRepo.FindByID(ref.StudentID)
	lecturer, errL := s.lecturerRepo.FindByUserID(verifierUserID)
	if errL != nil {
		return fmt.Errorf("lecturer profile not found")
	}

	if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
		return fmt.Errorf("forbidden: you are not the advisor")
	}

	return s.repo.Reject(id, note)
}

// =========================================================================
// 2. HANDLER METHODS (Berinteraksi dengan Fiber Ctx)
// =========================================================================

func (s *achievementService) GetAll(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	userID, _ := uuid.Parse(authData.UserID)

	var data []models.AchievementReference
	var err error

	switch authData.Role {
	case "Mahasiswa":
		student, _ := s.studentRepo.FindByUserID(userID)
		data, err = s.repo.FindReferencesByStudentID(student.ID)
	case "Dosen Wali":
		lecturer, _ := s.lecturerRepo.FindByUserID(userID)
		advisees, _ := s.lecturerRepo.FindAdvisees(lecturer.ID)
		var studentIDs []uuid.UUID
		for _, st := range advisees {
			studentIDs = append(studentIDs, st.ID)
		}
		data, err = s.repo.FindReferencesByStudentIDs(studentIDs)
	default:
		data, err = s.repo.FindAllReferences()
	}

	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Achievement list retrieved", data))
}

func (s *achievementService) GetDetail(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	id, _ := uuid.Parse(c.Params("id"))

	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Achievement not found", nil))
	}

	allowed := false
	switch authData.Role {
	case "Admin":
		allowed = true
	case "Mahasiswa":
		student, _ := s.studentRepo.FindByUserID(uuid.MustParse(authData.UserID))
		if ref.StudentID == student.ID {
			allowed = true
		}
	case "Dosen Wali":
		lecturer, _ := s.lecturerRepo.FindByUserID(uuid.MustParse(authData.UserID))
		if ref.Student.AdvisorID != nil && *ref.Student.AdvisorID == lecturer.ID {
			allowed = true
		}
	}

	if !allowed {
		return c.Status(403).JSON(helper.APIResponse("error", "Forbidden access", nil))
	}

	mongoData, _ := s.repo.GetMongoDetail(ref.MongoAchievementID)
	return c.Status(200).JSON(helper.APIResponse("success", "Achievement detail", fiber.Map{
		"reference": ref,
		"details":   mongoData,
	}))
}

func (s *achievementService) Create(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	if authData.Role != "Mahasiswa" {
		return c.Status(403).JSON(helper.APIResponse("error", "Only students can create achievements", nil))
	}

	student, err := s.studentRepo.FindByUserID(uuid.MustParse(authData.UserID))
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Student profile not found", nil))
	}

	if student.AdvisorID == nil {
		return c.Status(403).JSON(helper.APIResponse("error", "Advisor (Dosen Wali) not assigned yet", nil))
	}

	var input models.Achievement
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "Invalid JSON input", nil))
	}

	result, err := s.CreateAchievement(input, student.ID)
	if err != nil {
		// [BARU] Memberikan respon spesifik untuk membenarkan input
		return c.Status(400).JSON(helper.APIResponse("error", "Validation Failed: "+err.Error(), nil))
	}

	return c.Status(201).JSON(helper.APIResponse("success", "Achievement created successfully", result))
}

func (s *achievementService) Update(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))

	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Not found", nil))
	}

	student, _ := s.studentRepo.FindByUserID(uuid.MustParse(authData.UserID))
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(helper.APIResponse("error", "Unauthorized", nil))
	}

	if ref.Status != models.StatusDraft && ref.Status != models.StatusRejected {
		return c.Status(400).JSON(helper.APIResponse("error", "Cannot update: Current status is "+string(ref.Status), nil))
	}

	var input models.Achievement
	c.BodyParser(&input)
	if err := s.repo.UpdateMongo(ref.MongoAchievementID, input); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Achievement updated", nil))
}

func (s *achievementService) Delete(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))

	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Not found", nil))
	}

	if authData.Role == "Mahasiswa" {
		student, _ := s.studentRepo.FindByUserID(uuid.MustParse(authData.UserID))
		if ref.StudentID != student.ID || ref.Status != models.StatusDraft {
			return c.Status(403).JSON(helper.APIResponse("error", "Cannot delete", nil))
		}
	}

	s.repo.SoftDelete(id)
	return c.Status(200).JSON(helper.APIResponse("success", "Achievement deleted", nil))
}

func (s *achievementService) Submit(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	userID, _ := uuid.Parse(authData.UserID)
	id, _ := uuid.Parse(c.Params("id"))

	student, _ := s.studentRepo.FindByUserID(userID)
	
	if err := s.SubmitAchievement(id, student.ID); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	
	return c.Status(200).JSON(helper.APIResponse("success", "Achievement submitted for verification", nil))
}

func (s *achievementService) Verify(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	id, _ := uuid.Parse(c.Params("id"))
	verifierUserID := uuid.MustParse(authData.UserID)

	if err := s.VerifyAchievement(id, verifierUserID); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Achievement verified", nil))
}

func (s *achievementService) Reject(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	id, _ := uuid.Parse(c.Params("id"))
	verifierUserID := uuid.MustParse(authData.UserID)

	var input struct {
		Note string `json:"note"`
	}
	c.BodyParser(&input)

	if err := s.RejectAchievement(id, verifierUserID, input.Note); err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", err.Error(), nil))
	}

	return c.Status(200).JSON(helper.APIResponse("success", "Achievement rejected", nil))
}

func (s *achievementService) GetHistory(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	ref, err := s.repo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(helper.APIResponse("error", "Not found", nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "History retrieved", fiber.Map{
		"current_status": ref.Status,
		"submitted_at":   ref.SubmittedAt,
		"verified_at":    ref.VerifiedAt,
		"rejected_note":  ref.RejectionNote,
	}))
}

func (s *achievementService) AddAttachment(c *fiber.Ctx) error {
	authData, _ := middleware.CheckAuth(c.Get("Authorization"))
	id, _ := uuid.Parse(c.Params("id"))
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(helper.APIResponse("error", "File is required", nil))
	}

	ref, _ := s.repo.FindReferenceByID(id)
	student, _ := s.studentRepo.FindByUserID(uuid.MustParse(authData.UserID))
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(helper.APIResponse("error", "Unauthorized", nil))
	}

	uniqueName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	savePath := fmt.Sprintf("./uploads/%s", uniqueName)
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", "Failed to save file", nil))
	}

	attachment := models.Attachment{
		FileName:   file.Filename,
		FileURL:    "/uploads/" + uniqueName,
		FileType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	s.repo.AddAttachment(ref.MongoAchievementID, attachment)
	return c.Status(200).JSON(helper.APIResponse("success", "File uploaded", attachment))
}