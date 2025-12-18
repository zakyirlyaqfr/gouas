package route

import (
	"gouas/app/service"
	"gouas/middleware"

	"github.com/gofiber/fiber/v2"
)

func InitRoutes(
	app *fiber.App,
	authSvc service.AuthService,
	adminSvc service.AdminService,
	achSvc service.AchievementService,
	studentSvc service.StudentService,
	lecturerSvc service.LecturerService,
	reportSvc service.ReportService,
) {
	api := app.Group("/api/v1")

	// =========================================================================
	// 5.1 AUTHENTICATION
	// =========================================================================
	auth := api.Group("/auth")
	auth.Post("/login", authSvc.Login)
	auth.Post("/refresh", authSvc.Refresh)
	auth.Post("/logout", authSvc.Logout)
	auth.Get("/profile", authSvc.GetProfile)

	// =========================================================================
	// 5.2 USERS (ADMIN)
	// =========================================================================
	users := api.Group("/users")
	// Middleware Check: Hanya Admin
	users.Use(func(c *fiber.Ctx) error {
		authData, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil || authData.Role != "Admin" {
			return c.Status(403).JSON(fiber.Map{"status": "error", "message": "Forbidden"})
		}
		return c.Next()
	})

	users.Get("/", adminSvc.GetAllUsers)
	users.Get("/:id", adminSvc.GetUserDetail)
	users.Post("/", adminSvc.CreateUser)
	users.Put("/:id", adminSvc.UpdateUser)
	users.Delete("/:id", adminSvc.DeleteUser)
	users.Put("/:id/role", adminSvc.AssignRole)

	// =========================================================================
	// 5.4 ACHIEVEMENTS
	// =========================================================================
	ach := api.Group("/achievements")
	
	// Middleware Auth (Basic Check)
	ach.Use(func(c *fiber.Ctx) error {
		_, err := middleware.CheckAuth(c.Get("Authorization"))
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"status": "error", "message": "Unauthorized"})
		}
		return c.Next()
	})

	ach.Get("/", achSvc.GetAll)
	ach.Get("/:id", achSvc.GetDetail)
	ach.Post("/", achSvc.Create)
	ach.Put("/:id", achSvc.Update)
	ach.Delete("/:id", achSvc.Delete)
	ach.Post("/:id/submit", achSvc.Submit)
	ach.Post("/:id/verify", achSvc.Verify)
	ach.Post("/:id/reject", achSvc.Reject)
	ach.Get("/:id/history", achSvc.GetHistory)
	ach.Post("/:id/attachments", achSvc.AddAttachment)

	// =========================================================================
	// 5.5 STUDENTS & LECTURERS
	// =========================================================================
	api.Get("/students", studentSvc.GetAll)
	api.Get("/students/:id", studentSvc.GetDetail)
	api.Get("/students/:id/achievements", studentSvc.GetStudentAchievements)
	api.Put("/students/:id/advisor", studentSvc.AssignAdvisor)

	api.Get("/lecturers", lecturerSvc.GetAll)
	api.Get("/lecturers/:id/advisees", lecturerSvc.GetAdvisees)

	// =========================================================================
	// 5.8 REPORTS
	// =========================================================================
	api.Get("/reports/statistics", reportSvc.GetStatistics)
	api.Get("/reports/student/:id", reportSvc.GetStudentReport)
}