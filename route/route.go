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
	AdvisorID string `json:"advisor_id"` // ID dari tabel lecturers
}

type RejectRequest struct {
	Note string `json:"note"`
}

// ================= ROUTE SETUP =================

func SetupRoutes(app *fiber.App) {
	// 1. Init Dependencies
	authRepo := repository.NewAuthRepository()
	userRepo := repository.NewUserRepository()
	achRepo := repository.NewAchievementRepository()

	// Service Initialization
	authService := service.NewAuthService(authRepo)
	userService := service.NewUserService(userRepo, authRepo)
	achService := service.NewAchievementService(achRepo)

	// 2. Group API
	api := app.Group("/api/v1")

	// ================= AUTH ROUTES (PUBLIC) =================

	// POST /api/v1/auth/register (Self Register)
	api.Post("/auth/register", func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		roleUUID, _ := uuid.Parse(req.RoleID)
		user, err := authService.Register(req.Username, req.Email, req.Password, req.FullName, roleUUID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "User registered successfully", user)
	})

	// POST /api/v1/auth/login
	api.Post("/auth/login", func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		token, user, err := authService.Login(req.Username, req.Password)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
		}

		return helper.SuccessResponse(c, "Login success", fiber.Map{
			"token": token,
			"user":  user,
		})
	})

	// ================= USER PROFILE ROUTES (PROTECTED) =================
	
	userRoutes := api.Group("/users", middleware.Protected())

	// GET /api/v1/users/profile
	userRoutes.Get("/profile", func(c *fiber.Ctx) error {
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userIDStr := claims["user_id"].(string)
		userID, _ := uuid.Parse(userIDStr)

		user, err := authService.GetProfile(userID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "User not found")
		}
		return helper.SuccessResponse(c, "User profile retrieved", user)
	})

	// ================= ACHIEVEMENT ROUTES (MAHASISWA & DOSEN) =================
	achievements := api.Group("/achievements", middleware.Protected())

	// --- MAHASISWA FEATURES ---

	// POST /api/v1/achievements (Create Draft)
	achievements.Post("/", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		var req model.MongoAchievement
		// Parsing JSON Body (untuk field dinamis details dsb)
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid body")
		}

		// Ambil User ID dari Token
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userID, _ := uuid.Parse(claims["user_id"].(string))

		result, err := achService.CreateDraft(userID, req)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Achievement created (draft)", result)
	})

	// GET /api/v1/achievements (List Own Achievements)
	achievements.Get("/", middleware.PermissionCheck("achievement:read"), func(c *fiber.Ctx) error {
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userID, _ := uuid.Parse(claims["user_id"].(string))

		results, err := achService.GetMyAchievements(userID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "My achievements", results)
	})

	// POST /api/v1/achievements/:id/submit (Submit Prestasi)
	achievements.Post("/:id/submit", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		achID, _ := uuid.Parse(idStr)

		if err := achService.SubmitAchievement(achID); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Achievement submitted successfully", nil)
	})
	
	// DELETE /api/v1/achievements/:id (Delete Draft)
	achievements.Delete("/:id", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		achID, _ := uuid.Parse(idStr)

		if err := achService.DeleteAchievement(achID); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Achievement deleted successfully", nil)
	})

	// --- DOSEN WALI FEATURES (TAHAP 6) ---

	// GET /api/v1/achievements/advisees (List Prestasi Mahasiswa Bimbingan)
	achievements.Get("/advisees", middleware.PermissionCheck("achievement:verify"), func(c *fiber.Ctx) error {
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userID, _ := uuid.Parse(claims["user_id"].(string))

		results, err := achService.GetAdviseeAchievements(userID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Advisee achievements retrieved", results)
	})

	// POST /api/v1/achievements/:id/verify (Verify Prestasi)
	achievements.Post("/:id/verify", middleware.PermissionCheck("achievement:verify"), func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		achID, _ := uuid.Parse(idStr)
		
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userID, _ := uuid.Parse(claims["user_id"].(string))

		if err := achService.VerifyAchievement(userID, achID); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		}
		return helper.SuccessResponse(c, "Achievement verified successfully", nil)
	})

	// POST /api/v1/achievements/:id/reject (Reject Prestasi)
	achievements.Post("/:id/reject", middleware.PermissionCheck("achievement:verify"), func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		achID, _ := uuid.Parse(idStr)

		var req RejectRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Note is required")
		}

		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userID, _ := uuid.Parse(claims["user_id"].(string))

		if err := achService.RejectAchievement(userID, achID, req.Note); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		}
		return helper.SuccessResponse(c, "Achievement rejected", nil)
	})

	// ================= ADMIN USER MANAGEMENT ROUTES (RBAC) =================
	
	admin := api.Group("/users", middleware.Protected(), middleware.PermissionCheck("user:manage"))

	// POST /api/v1/users (Create User by Admin)
	admin.Post("/", func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		roleUUID, _ := uuid.Parse(req.RoleID)
		user, err := userService.CreateUser(req.Username, req.Email, req.Password, req.FullName, roleUUID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "User created by admin", user)
	})

	// GET /api/v1/users (List All Users)
	admin.Get("/", func(c *fiber.Ctx) error {
		users, err := userService.GetAllUsers()
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "List of all users", users)
	})

	// PUT /api/v1/users/:id/role (Change User Role)
	admin.Put("/:id/role", func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		userID, err := uuid.Parse(idStr)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid User UUID")
		}

		var req ChangeRoleRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		roleID, err := uuid.Parse(req.RoleID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Role UUID")
		}

		if err := userService.ChangeRole(userID, roleID); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "User role updated successfully", nil)
	})

	// POST /api/v1/users/students (Setup Student Profile)
	admin.Post("/students", func(c *fiber.Ctx) error {
		var req SetupStudentRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		userUUID, err := uuid.Parse(req.UserID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid User UUID")
		}

		if err := userService.SetupStudentProfile(userUUID, req.NIM, req.ProgramStudy, req.AcademicYear); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Student profile setup successfully", nil)
	})

	// POST /api/v1/users/lecturers (Setup Lecturer Profile)
	admin.Post("/lecturers", func(c *fiber.Ctx) error {
		var req SetupLecturerRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		userUUID, err := uuid.Parse(req.UserID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid User UUID")
		}

		if err := userService.SetupLecturerProfile(userUUID, req.NIP, req.Department); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Lecturer profile setup successfully", nil)
	})

	// PUT /api/v1/users/students/:id/advisor (Assign Advisor)
	// Note: :id disini merujuk pada STUDENT ID (UUID di tabel students), bukan User ID.
	admin.Put("/students/:id/advisor", func(c *fiber.Ctx) error {
		studentIDStr := c.Params("id")
		studentID, err := uuid.Parse(studentIDStr)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Student UUID")
		}

		var req AssignAdvisorRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
		}

		advisorID, err := uuid.Parse(req.AdvisorID) // ID dari tabel lecturers
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Advisor/Lecturer UUID")
		}

		if err := userService.AssignAdvisor(studentID, advisorID); err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		return helper.SuccessResponse(c, "Advisor assigned successfully", nil)
	})
}