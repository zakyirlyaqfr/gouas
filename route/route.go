package route

import (
	"gouas/app/models"
	"gouas/app/service"
	"gouas/helper"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func InitRoutes(
	app *fiber.App,
	authService service.AuthService,
	adminService service.AdminService,
	achievementService service.AchievementService,
	studentService service.StudentService,
	lecturerService service.LecturerService,
	reportService service.ReportService,
) {
	api := app.Group("/api/v1")

	// Helper wrapper agar lebih ringkas saat dipakai di handler bawah
	jsonResponse := func(c *fiber.Ctx, code int, status string, message string, data interface{}) error {
		// Kita panggil helper.APIResponse untuk dapatkan struct standar
		response := helper.APIResponse(status, message, data)
		// Fiber otomatis serialize struct ke JSON
		return c.Status(code).JSON(response)
	}

	// --- 5.1 AUTHENTICATION ---
	auth := api.Group("/auth")
	auth.Post("/login", func(c *fiber.Ctx) error {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&input); err != nil {
			return jsonResponse(c, 400, "error", "Invalid input", nil)
		}
		token, err := authService.Login(input.Username, input.Password)
		if err != nil {
			return jsonResponse(c, 401, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Login successful", map[string]string{"token": token})
	})

	// --- 5.2 USERS (ADMIN) ---
	users := api.Group("/users")
	users.Post("/", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || authData.Role != "Admin" {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}

		var input struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
			FullName string `json:"fullName"`
			RoleName string `json:"roleName"`
		}
		if err := c.BodyParser(&input); err != nil {
			return jsonResponse(c, 400, "error", "Invalid input", nil)
		}

		user, err := adminService.CreateUser(input.Username, input.Email, input.Password, input.FullName, input.RoleName)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 201, "success", "User created", user)
	})

	// --- 5.4 ACHIEVEMENTS ---
	achievements := api.Group("/achievements")

	achievements.Post("/", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		if !middleware.HasPermission(authData.Permissions, "achievement:create") {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}

		var input models.Achievement
		if err := c.BodyParser(&input); err != nil {
			return jsonResponse(c, 400, "error", "Invalid JSON", nil)
		}

		studentID, _ := uuid.Parse(authData.UserID)
		result, err := achievementService.Create(studentID, input)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 201, "success", "Achievement created", result)
	})

	achievements.Post("/:id/submit", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		achID, _ := uuid.Parse(c.Params("id"))
		userID, _ := uuid.Parse(authData.UserID)

		if err := achievementService.Submit(achID, userID); err != nil {
			return jsonResponse(c, 400, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement submitted", nil)
	})

	achievements.Post("/:id/verify", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || !middleware.HasPermission(authData.Permissions, "achievement:verify") {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}

		achID, _ := uuid.Parse(c.Params("id"))
		lecturerID, _ := uuid.Parse(authData.UserID)

		if err := achievementService.Verify(achID, lecturerID); err != nil {
			return jsonResponse(c, 400, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement verified", nil)
	})

	achievements.Post("/:id/reject", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || !middleware.HasPermission(authData.Permissions, "achievement:verify") {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}

		var input struct {
			Note string `json:"note"`
		}
		c.BodyParser(&input)

		achID, _ := uuid.Parse(c.Params("id"))
		lecturerID, _ := uuid.Parse(authData.UserID)

		if err := achievementService.Reject(achID, lecturerID, input.Note); err != nil {
			return jsonResponse(c, 400, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement rejected", nil)
	})

	// --- 5.5 STUDENTS ---
	api.Get("/students", func(c *fiber.Ctx) error {
		students, err := studentService.GetAll()
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Student list", students)
	})

	// --- 5.8 REPORTS ---
	api.Get("/reports/statistics", func(c *fiber.Ctx) error {
		stats, err := reportService.GetStatistics()
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Statistics", stats)
	})
}