package route

import (
	"fmt" // [BARU]
	"gouas/app/models"
	"gouas/app/service"
	"gouas/helper"
	"gouas/middleware"
	"os"   // [BARU]
	"time" // [BARU]

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

	// Helper wrapper agar response konsisten JSON
	jsonResponse := func(c *fiber.Ctx, code int, status string, message string, data interface{}) error {
		response := helper.APIResponse(status, message, data)
		return c.Status(code).JSON(response)
	}

	// =========================================================================
	// 5.1 AUTHENTICATION
	// =========================================================================
	// ... code sebelumnya ...
	auth := api.Group("/auth")

	auth.Post("/login", func(c *fiber.Ctx) error {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&input); err != nil {
			return jsonResponse(c, 400, "error", "Invalid input", nil)
		}

		// Panggil service yang sekarang mengembalikan 2 token
		accessToken, refreshToken, err := authService.Login(input.Username, input.Password)
		if err != nil {
			return jsonResponse(c, 401, "error", err.Error(), nil)
		}

		return jsonResponse(c, 200, "success", "Login successful", fiber.Map{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		})
	})

	auth.Post("/refresh", func(c *fiber.Ctx) error {
		var input struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.BodyParser(&input); err != nil {
			return jsonResponse(c, 400, "error", "Invalid input", nil)
		}

		// [PERBAIKAN] Service Refresh sekarang hanya mengembalikan (NewAccessToken, Error)
		// Karena Refresh Token ID tidak berubah di DB (tetap valid 24 jam)
		newAccess, err := authService.Refresh(input.RefreshToken)
		if err != nil {
			return jsonResponse(c, 401, "error", err.Error(), nil)
		}

		return jsonResponse(c, 200, "success", "Token refreshed", fiber.Map{
			"accessToken":  newAccess,
			"refreshToken": input.RefreshToken, // Kita kembalikan token lama karena masih valid
		})
	})
	// ... sisanya sama ...

	auth.Post("/logout", func(c *fiber.Ctx) error {
		// Ambil User ID dari token yang sedang login
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		userID, _ := uuid.Parse(authData.UserID)

		// Panggil Service Logout (Set NULL di DB)
		if err := authService.Logout(userID); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}

		return jsonResponse(c, 200, "success", "Logged out successfully (All tokens revoked)", nil)
	})

	auth.Get("/profile", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}
		return jsonResponse(c, 200, "success", "User Profile", authData)
	})

	// =========================================================================
	// 5.2 USERS (ADMIN)
	// =========================================================================
	users := api.Group("/users")

	// Middleware Check: Hanya Admin
	users.Use(func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || authData.Role != "Admin" {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}
		return c.Next()
	})

	users.Get("/", func(c *fiber.Ctx) error {
		users, err := adminService.GetAllUsers()
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "User list", users)
	})

	users.Get("/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		user, err := adminService.GetUserDetail(id)
		if err != nil {
			return jsonResponse(c, 404, "error", "User not found", nil)
		}
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
		if err := c.BodyParser(&input); err != nil {
			return jsonResponse(c, 400, "error", "Invalid input", nil)
		}

		// Create User sekaligus Create Profile (Student/Lecturer) otomatis di Service
		user, err := adminService.CreateUser(input.Username, input.Email, input.Password, input.FullName, input.RoleName)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 201, "success", "User created", user)
	})

	users.Put("/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		var input struct {
			FullName string `json:"fullName"`
			Email    string `json:"email"`
		}
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
		var input struct {
			RoleName string `json:"roleName"`
		}
		c.BodyParser(&input)
		if err := adminService.AssignRole(id, input.RoleName); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Role updated", nil)
	})

	// =========================================================================
	// 5.4 ACHIEVEMENTS
	// =========================================================================
	ach := api.Group("/achievements")

	ach.Get("/", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		// Filter logic: Mahasiswa lihat punya sendiri, Admin/Dosen lihat semua
		userID, _ := uuid.Parse(authData.UserID)

		// Jika mahasiswa, kita perlu mencari StudentID-nya dulu
		var targetID uuid.UUID
		if authData.Role == "Mahasiswa" {
			student, err := studentService.GetProfileByUserID(userID)
			if err != nil {
				return jsonResponse(c, 404, "error", "Student profile not found", nil)
			}
			targetID = student.ID
		} else {
			targetID = userID // Untuk admin/dosen param ini diabaikan di service
		}

		data, err := achievementService.GetAll(authData.Role, targetID)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement list", data)
	})

	ach.Get("/:id", func(c *fiber.Ctx) error {
		_, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		id, _ := uuid.Parse(c.Params("id"))
		data, err := achievementService.GetDetail(id)
		if err != nil {
			return jsonResponse(c, 404, "error", "Not found", nil)
		}
		return jsonResponse(c, 200, "success", "Detail", data)
	})

	ach.Post("/", func(c *fiber.Ctx) error {
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

		// Cari Student Profile berdasarkan UserID
		userID, _ := uuid.Parse(authData.UserID)
		student, err := studentService.GetProfileByUserID(userID)
		if err != nil {
			return jsonResponse(c, 404, "error", "Student profile not found. Contact admin.", nil)
		}

		// Gunakan StudentID asli
		result, err := achievementService.Create(student.ID, input)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 201, "success", "Achievement created", result)
	})

	ach.Put("/:id", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		id, _ := uuid.Parse(c.Params("id"))
		userID, _ := uuid.Parse(authData.UserID)

		// Cari student ID user yang login
		student, err := studentService.GetProfileByUserID(userID)
		if err != nil {
			return jsonResponse(c, 404, "error", "Profile not found", nil)
		}

		var input models.Achievement
		c.BodyParser(&input)

		if err := achievementService.Update(id, student.ID, input); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement updated", nil)
	})

	ach.Delete("/:id", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		id, _ := uuid.Parse(c.Params("id"))
		userID, _ := uuid.Parse(authData.UserID)

		student, err := studentService.GetProfileByUserID(userID)
		if err != nil {
			return jsonResponse(c, 404, "error", "Profile not found", nil)
		}

		if err := achievementService.Delete(id, student.ID); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Achievement deleted", nil)
	})

	ach.Post("/:id/submit", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		achID, _ := uuid.Parse(c.Params("id"))
		userID, _ := uuid.Parse(authData.UserID)

		student, err := studentService.GetProfileByUserID(userID)
		if err != nil {
			return jsonResponse(c, 404, "error", "Profile not found", nil)
		}

		if err := achievementService.Submit(achID, student.ID); err != nil {
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

	ach.Get("/:id/history", func(c *fiber.Ctx) error {
		_, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}
		id, _ := uuid.Parse(c.Params("id"))
		history, err := achievementService.GetHistory(id)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "History", history)
	})

	// --- [PERBAIKAN] Endpoint Upload File Fisik ---
	ach.Post("/:id/attachments", func(c *fiber.Ctx) error {
		// 1. Cek Auth
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return jsonResponse(c, 401, "error", "Unauthorized", nil)
		}

		// 2. Cek Profile Mahasiswa
		id, _ := uuid.Parse(c.Params("id"))
		userID, _ := uuid.Parse(authData.UserID)
		student, err := studentService.GetProfileByUserID(userID)
		if err != nil {
			return jsonResponse(c, 404, "error", "Profile not found", nil)
		}

		// 3. Ambil File dari Form-Data (Key: "file")
		file, err := c.FormFile("file")
		if err != nil {
			return jsonResponse(c, 400, "error", "File is required (key: 'file')", nil)
		}

		// 4. Generate Nama Unik & Path
		uniqueName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
		savePath := fmt.Sprintf("./uploads/%s", uniqueName)
		fileURL := fmt.Sprintf("/uploads/%s", uniqueName) // Ini URL public

		// 5. Simpan File ke Disk
		if err := c.SaveFile(file, savePath); err != nil {
			return jsonResponse(c, 500, "error", "Failed to save file: "+err.Error(), nil)
		}

		// 6. Ambil Caption (jika ada)
		caption := c.FormValue("caption")
		if caption == "" {
			caption = file.Filename
		}

		// 7. Simpan Info ke Database via Service
		if err := achievementService.AddAttachment(id, student.ID, caption, fileURL); err != nil {
			os.Remove(savePath) // Hapus file jika DB gagal
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}

		return jsonResponse(c, 200, "success", "Attachment added", fiber.Map{
			"fileUrl": fileURL,
		})
	})
	// ----------------------------------------------

	// =========================================================================
	// 5.5 STUDENTS & LECTURERS
	// =========================================================================
	api.Get("/students", func(c *fiber.Ctx) error {
		students, err := studentService.GetAll()
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Student list", students)
	})

	api.Get("/students/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		student, err := studentService.GetDetail(id)
		if err != nil {
			return jsonResponse(c, 404, "error", "Not found", nil)
		}
		return jsonResponse(c, 200, "success", "Student detail", student)
	})

	api.Get("/students/:id/achievements", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		data, err := achievementService.GetAll("Mahasiswa", id)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Student achievements", data)
	})

	api.Put("/students/:id/advisor", func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || authData.Role != "Admin" {
			return jsonResponse(c, 403, "error", "Forbidden", nil)
		}

		id, _ := uuid.Parse(c.Params("id"))
		var input struct {
			AdvisorID string `json:"advisorId"`
		}
		c.BodyParser(&input)
		advID, _ := uuid.Parse(input.AdvisorID)

		if err := studentService.AssignAdvisor(id, advID); err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Advisor assigned", nil)
	})

	api.Get("/lecturers", func(c *fiber.Ctx) error {
		lecturers, err := lecturerService.GetAll()
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Lecturer list", lecturers)
	})

	api.Get("/lecturers/:id/advisees", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		students, err := lecturerService.GetAdvisees(id)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Advisees list", students)
	})

	// =========================================================================
	// 5.8 REPORTS
	// =========================================================================
	api.Get("/reports/statistics", func(c *fiber.Ctx) error {
		stats, err := reportService.GetStatistics()
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Statistics", stats)
	})

	api.Get("/reports/student/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		// Reuse logic: fetch achievements as report
		data, err := achievementService.GetAll("Mahasiswa", id)
		if err != nil {
			return jsonResponse(c, 500, "error", err.Error(), nil)
		}
		return jsonResponse(c, 200, "success", "Student Report", data)
	})
}
