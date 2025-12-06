package route

import (
	"gouas/app/model"
	"gouas/app/repository"
	"gouas/app/service"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ================= STRUCTS REQUEST BODY =================

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	RoleID   string `json:"role_id"`
}

type ChangeRoleRequest struct {
	RoleID string `json:"role_id"`
}

type SetupStudentRequest struct {
	UserID       string `json:"user_id"`
	NIM          string `json:"nim"`
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
}

type SetupLecturerRequest struct {
	UserID     string `json:"user_id"`
	NIP        string `json:"nip"`
	Department string `json:"department"`
}

type AssignAdvisorRequest struct {
	AdvisorID string `json:"advisor_id"`
}

type RejectRequest struct {
	Note string `json:"note"`
}

// ================= HANDLER STRUCT =================
// WebHandler membungkus semua service agar bisa diakses oleh method handler
type WebHandler struct {
	authService   service.AuthService
	userService   service.UserService
	achService    service.AchievementService
	reportService service.ReportService
}

// ================= HANDLER METHODS (WITH SWAGGER ANNOTATIONS) =================

// Register godoc
// @Summary      Register User
// @Description  Mendaftarkan user baru
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Payload Register"
// @Success      200 {object} helper.APIResponse
// @Failure      400 {object} helper.APIResponse
// @Router       /auth/register [post]
func (h *WebHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}
	roleUUID, _ := uuid.Parse(req.RoleID)
	user, err := h.authService.Register(req.Username, req.Email, req.Password, req.FullName, roleUUID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.SuccessResponse(c, "User registered successfully", user)
}

// Login godoc
// @Summary      Login User
// @Description  Masuk ke sistem dan mendapatkan JWT Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Payload Login"
// @Success      200 {object} helper.APIResponse
// @Failure      401 {object} helper.APIResponse
// @Router       /auth/login [post]
func (h *WebHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}
	token, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	return helper.SuccessResponse(c, "Login success", fiber.Map{"token": token, "user": user})
}

// GetProfile godoc
// @Summary      Get My Profile
// @Description  Melihat profil user yang sedang login
// @Tags         User
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} helper.APIResponse
// @Router       /users/profile [get]
func (h *WebHandler) GetProfile(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))
	user, err := h.authService.GetProfile(userID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "User not found")
	}
	return helper.SuccessResponse(c, "User profile retrieved", user)
}

// CreateAchievement godoc
// @Summary      Create Achievement Draft
// @Description  Mahasiswa membuat laporan prestasi baru (Draft)
// @Tags         Achievements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.MongoAchievement true "Data Prestasi"
// @Success      200 {object} helper.APIResponse
// @Router       /achievements [post]
func (h *WebHandler) CreateAchievement(c *fiber.Ctx) error {
	var req model.MongoAchievement
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid body")
	}
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))
	result, err := h.achService.CreateDraft(userID, req)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.SuccessResponse(c, "Achievement created (draft)", result)
}

// GetMyAchievements godoc
// @Summary      List My Achievements
// @Description  Melihat daftar prestasi milik sendiri
// @Tags         Achievements
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} helper.APIResponse
// @Router       /achievements [get]
func (h *WebHandler) GetMyAchievements(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))
	results, err := h.achService.GetMyAchievements(userID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.SuccessResponse(c, "My achievements", results)
}

// SubmitAchievement godoc
// @Summary      Submit Achievement
// @Description  Mengubah status prestasi dari draft ke submitted
// @Tags         Achievements
// @Security     BearerAuth
// @Param        id path string true "Achievement ID (UUID)"
// @Success      200 {object} helper.APIResponse
// @Router       /achievements/{id}/submit [post]
func (h *WebHandler) SubmitAchievement(c *fiber.Ctx) error {
	idStr := c.Params("id")
	achID, _ := uuid.Parse(idStr)
	if err := h.achService.SubmitAchievement(achID); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.SuccessResponse(c, "Achievement submitted", nil)
}

// DeleteAchievement godoc
// @Summary      Delete Achievement Draft
// @Description  Menghapus (soft delete) prestasi draft
// @Tags         Achievements
// @Security     BearerAuth
// @Param        id path string true "Achievement ID (UUID)"
// @Success      200 {object} helper.APIResponse
// @Router       /achievements/{id} [delete]
func (h *WebHandler) DeleteAchievement(c *fiber.Ctx) error {
	idStr := c.Params("id")
	achID, _ := uuid.Parse(idStr)
	if err := h.achService.DeleteAchievement(achID); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.SuccessResponse(c, "Achievement deleted", nil)
}

// GetAdviseesAchievements godoc
// @Summary      Get Advisee Achievements
// @Description  Dosen melihat daftar prestasi mahasiswa bimbingan
// @Tags         Verification
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} helper.APIResponse
// @Router       /achievements/advisees [get]
func (h *WebHandler) GetAdviseesAchievements(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))

	results, err := h.achService.GetAdviseeAchievements(userID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.SuccessResponse(c, "Advisee achievements", results)
}

// VerifyAchievement godoc
// @Summary      Verify Achievement
// @Description  Dosen memverifikasi prestasi
// @Tags         Verification
// @Security     BearerAuth
// @Param        id path string true "Achievement ID (UUID)"
// @Success      200 {object} helper.APIResponse
// @Router       /achievements/{id}/verify [post]
func (h *WebHandler) VerifyAchievement(c *fiber.Ctx) error {
	idStr := c.Params("id")
	achID, _ := uuid.Parse(idStr)
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))
	if err := h.achService.VerifyAchievement(userID, achID); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return helper.SuccessResponse(c, "Achievement verified", nil)
}

// RejectAchievement godoc
// @Summary      Reject Achievement
// @Description  Dosen menolak prestasi dengan catatan
// @Tags         Verification
// @Security     BearerAuth
// @Param        id path string true "Achievement ID (UUID)"
// @Param        request body RejectRequest true "Alasan Penolakan"
// @Success      200 {object} helper.APIResponse
// @Router       /achievements/{id}/reject [post]
func (h *WebHandler) RejectAchievement(c *fiber.Ctx) error {
	idStr := c.Params("id")
	achID, _ := uuid.Parse(idStr)
	var req RejectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid body")
	}
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))
	if err := h.achService.RejectAchievement(userID, achID, req.Note); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return helper.SuccessResponse(c, "Achievement rejected", nil)
}

// GetDashboardStats godoc
// @Summary      Dashboard Statistics
// @Description  Melihat statistik prestasi
// @Tags         Reports
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} helper.APIResponse
// @Router       /reports/statistics [get]
func (h *WebHandler) GetDashboardStats(c *fiber.Ctx) error {
	stats, err := h.reportService.GetDashboardStatistics()
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.SuccessResponse(c, "Dashboard statistics", stats)
}

// ================= SETUP ROUTING =================

func SetupRoutes(app *fiber.App) {
	// 1. Init Dependencies
	authRepo := repository.NewAuthRepository()
	userRepo := repository.NewUserRepository()
	achRepo := repository.NewAchievementRepository()
	reportRepo := repository.NewReportRepository()

	// 2. Init Handler Struct
	h := &WebHandler{
		authService:   service.NewAuthService(authRepo),
		userService:   service.NewUserService(userRepo, authRepo),
		achService:    service.NewAchievementService(achRepo),
		reportService: service.NewReportService(reportRepo),
	}

	// 3. Register Routes
	api := app.Group("/api/v1")

	// Auth
	api.Post("/auth/register", h.Register)
	api.Post("/auth/login", h.Login)

	// Users
	userRoutes := api.Group("/users", middleware.Protected())
	userRoutes.Get("/profile", h.GetProfile)

	// Achievements
	ach := api.Group("/achievements", middleware.Protected())
	ach.Post("/", middleware.PermissionCheck("achievement:create"), h.CreateAchievement)
	ach.Get("/", middleware.PermissionCheck("achievement:read"), h.GetMyAchievements)
	ach.Post("/:id/submit", middleware.PermissionCheck("achievement:create"), h.SubmitAchievement)
	ach.Delete("/:id", middleware.PermissionCheck("achievement:create"), h.DeleteAchievement)
	
	// Verification
	ach.Get("/advisees", middleware.PermissionCheck("achievement:verify"), h.GetAdviseesAchievements)
	ach.Post("/:id/verify", middleware.PermissionCheck("achievement:verify"), h.VerifyAchievement)
	ach.Post("/:id/reject", middleware.PermissionCheck("achievement:verify"), h.RejectAchievement)

	// Reports
	api.Get("/reports/statistics", middleware.Protected(), h.GetDashboardStats)

	// ================= ADMIN ROUTES (INLINE) =================
	// Admin routes dibiarkan inline karena fokus Swagger biasanya untuk client API (Student/Lecturer)
	// Jika ingin muncul di Swagger juga, buatkan method di WebHandler seperti di atas.
	
	admin := api.Group("/users", middleware.Protected(), middleware.PermissionCheck("user:manage"))

	// Create User
	admin.Post("/", func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}
		roleUUID, _ := uuid.Parse(req.RoleID)
		user, err := h.userService.CreateUser(req.Username, req.Email, req.Password, req.FullName, roleUUID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "User created by admin", user)
	})

	// List Users
	admin.Get("/", func(c *fiber.Ctx) error {
		users, err := h.userService.GetAllUsers()
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "List of all users", users)
	})

	// Change Role
	admin.Put("/:id/role", func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		userID, _ := uuid.Parse(idStr)
		var req ChangeRoleRequest
		if err := c.BodyParser(&req); err != nil { return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid body") }
		roleID, _ := uuid.Parse(req.RoleID)
		
		if err := h.userService.ChangeRole(userID, roleID); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "User role updated", nil)
	})

	// Setup Student
	admin.Post("/students", func(c *fiber.Ctx) error {
		var req SetupStudentRequest
		if err := c.BodyParser(&req); err != nil { return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid body") }
		userUUID, _ := uuid.Parse(req.UserID)
		if err := h.userService.SetupStudentProfile(userUUID, req.NIM, req.ProgramStudy, req.AcademicYear); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Student profile setup", nil)
	})

	// Setup Lecturer
	admin.Post("/lecturers", func(c *fiber.Ctx) error {
		var req SetupLecturerRequest
		if err := c.BodyParser(&req); err != nil { return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid body") }
		userUUID, _ := uuid.Parse(req.UserID)
		if err := h.userService.SetupLecturerProfile(userUUID, req.NIP, req.Department); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Lecturer profile setup", nil)
	})

	// Assign Advisor
	admin.Put("/students/:id/advisor", func(c *fiber.Ctx) error {
		studentIDStr := c.Params("id")
		studentID, _ := uuid.Parse(studentIDStr)
		var req AssignAdvisorRequest
		if err := c.BodyParser(&req); err != nil { return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid body") }
		advisorID, _ := uuid.Parse(req.AdvisorID)
		
		if err := h.userService.AssignAdvisor(studentID, advisorID); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Advisor assigned", nil)
	})
}