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

	jsonResponse := func(c *fiber.Ctx, code int, status string, message string, data interface{}) error {
		response := helper.APIResponse(status, message, data)
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
	
	auth.Post("/refresh", func(c *fiber.Ctx) error {
		return jsonResponse(c, 200, "success", "Token refreshed (mock)", nil)
	})

	auth.Post("/logout", func(c *fiber.Ctx) error {
		return jsonResponse(c, 200, "success", "Logged out successfully", nil)
	})

	auth.Get("/profile", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		return jsonResponse(c, 200, "success", "User Profile", authData)
	})


	// --- 5.2 USERS (ADMIN) ---
	users := api.Group("/users")
	users.Use(func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || authData.Role != "Admin" {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}
		return c.Next()
	})

	users.Get("/", func(c *fiber.Ctx) error {
		users, err := adminService.GetAllUsers()
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "User list", users)
	})

	users.Get("/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		user, err := adminService.GetUserDetail(id)
		if err != nil { return jsonResponse(c, 404, "error", "User not found", nil) }
		return jsonResponse(c, 200, "success", "User detail", user)
	})

	users.Post("/", func(c *fiber.Ctx) error {
		var input struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
			FullName string `json:"fullName"`
			RoleName string `json:"roleName"`
		}
		if err := c.BodyParser(&input); err != nil { return jsonResponse(c, 400, "error", "Invalid input", nil) }
		user, err := adminService.CreateUser(input.Username, input.Email, input.Password, input.FullName, input.RoleName)
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 201, "success", "User created", user)
	})

	users.Put("/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		var input struct { FullName string `json:"fullName"`; Email string `json:"email"` }
		c.BodyParser(&input)
		if err := adminService.UpdateUser(id, input.FullName, input.Email); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "User updated", nil)
	})

	users.Delete("/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		if err := adminService.DeleteUser(id); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "User deleted", nil)
	})

	users.Put("/:id/role", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		var input struct { RoleName string `json:"roleName"` }
		c.BodyParser(&input)
		if err := adminService.AssignRole(id, input.RoleName); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Role updated", nil)
	})


	// --- 5.4 ACHIEVEMENTS ---
	ach := api.Group("/achievements")
	
	ach.Get("/", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		
		userID, _ := uuid.Parse(authData.UserID)
		data, err := achievementService.GetAll(authData.Role, userID)
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "Achievement list", data)
	})

	ach.Get("/:id", func(c *fiber.Ctx) error {
		// PERBAIKAN: authData diganti _ karena tidak dipakai variable-nya, hanya untuk cek error nil
		_, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		
		id, _ := uuid.Parse(c.Params("id"))
		data, err := achievementService.GetDetail(id)
		if err != nil { return jsonResponse(c, 404, "error", "Not found", nil) }
		return jsonResponse(c, 200, "success", "Detail", data)
	})

	ach.Post("/", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		if !middleware.HasPermission(authData.Permissions, "achievement:create") {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}

		var input models.Achievement
		if err := c.BodyParser(&input); err != nil { return jsonResponse(c, 400, "error", "Invalid JSON", nil) }
		studentID, _ := uuid.Parse(authData.UserID)
		result, err := achievementService.Create(studentID, input)
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 201, "success", "Achievement created", result)
	})

	ach.Put("/:id", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		
		id, _ := uuid.Parse(c.Params("id"))
		studentID, _ := uuid.Parse(authData.UserID)
		var input models.Achievement
		c.BodyParser(&input)

		if err := achievementService.Update(id, studentID, input); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement updated", nil)
	})

	ach.Delete("/:id", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		
		id, _ := uuid.Parse(c.Params("id"))
		studentID, _ := uuid.Parse(authData.UserID)
		if err := achievementService.Delete(id, studentID); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement deleted", nil)
	})

	ach.Post("/:id/submit", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		achID, _ := uuid.Parse(c.Params("id"))
		userID, _ := uuid.Parse(authData.UserID)
		if err := achievementService.Submit(achID, userID); err != nil {
			return jsonResponse(c, 400, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement submitted", nil)
	})

	ach.Post("/:id/verify", func(c *fiber.Ctx) error {
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

	ach.Post("/:id/reject", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || !middleware.HasPermission(authData.Permissions, "achievement:verify") {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}
		var input struct { Note string `json:"note"` }
		c.BodyParser(&input)
		achID, _ := uuid.Parse(c.Params("id"))
		lecturerID, _ := uuid.Parse(authData.UserID)
		if err := achievementService.Reject(achID, lecturerID, input.Note); err != nil {
			return jsonResponse(c, 400, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement rejected", nil)
	})

	ach.Get("/:id/history", func(c *fiber.Ctx) error {
		// PERBAIKAN: authData diganti _
		_, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		id, _ := uuid.Parse(c.Params("id"))
		history, err := achievementService.GetHistory(id)
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "History", history)
	})

	ach.Post("/:id/attachments", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil { return jsonResponse(c, 401, "error", "Unauthorized", nil) }
		
		id, _ := uuid.Parse(c.Params("id"))
		studentID, _ := uuid.Parse(authData.UserID)
		var input struct { FileName string `json:"fileName"`; FileUrl string `json:"fileUrl"` }
		c.BodyParser(&input)
		
		if err := achievementService.AddAttachment(id, studentID, input.FileName, input.FileUrl); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Attachment added", nil)
	})


	// --- 5.5 STUDENTS & LECTURERS ---
	api.Get("/students", func(c *fiber.Ctx) error {
		students, err := studentService.GetAll()
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "Student list", students)
	})

	api.Get("/students/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		student, err := studentService.GetDetail(id)
		if err != nil { return jsonResponse(c, 404, "error", "Not found", nil) }
		return jsonResponse(c, 200, "success", "Student detail", student)
	})

	api.Get("/students/:id/achievements", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		data, err := achievementService.GetAll("Mahasiswa", id) 
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "Student achievements", data)
	})

	api.Put("/students/:id/advisor", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || authData.Role != "Admin" { return jsonResponse(c, 403, "error", "Forbidden", nil) }

		id, _ := uuid.Parse(c.Params("id"))
		var input struct { AdvisorID string `json:"advisorId"` }
		c.BodyParser(&input)
		advID, _ := uuid.Parse(input.AdvisorID)
		
		if err := studentService.AssignAdvisor(id, advID); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Advisor assigned", nil)
	})

	api.Get("/lecturers", func(c *fiber.Ctx) error {
		lecturers, err := lecturerService.GetAll()
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "Lecturer list", lecturers)
	})

	api.Get("/lecturers/:id/advisees", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		students, err := lecturerService.GetAdvisees(id)
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "Advisees list", students)
	})


	// --- 5.8 REPORTS ---
	api.Get("/reports/statistics", func(c *fiber.Ctx) error {
		stats, err := reportService.GetStatistics()
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "Statistics", stats)
	})
	
	api.Get("/reports/student/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		data, err := achievementService.GetAll("Mahasiswa", id)
		if err != nil { return jsonResponse(c, 500, "error", err.Error(), nil) }
		return jsonResponse(c, 200, "success", "Student Report", data)
	})
}