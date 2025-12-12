package main

import (
	"fmt"
	"gouas/app/models"
	"gouas/app/repository"
	"gouas/app/service"
	"gouas/config"
	"gouas/database"
	_ "gouas/docs"
	"gouas/helper"
	"gouas/route"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

// Fungsi untuk seeding data awal (Role & Admin)
func seedDatabase(db *gorm.DB) {
	roles := []string{"Admin", "Mahasiswa", "Dosen Wali"}
	for _, r := range roles {
		var roleExist models.Role
		if err := db.Where("name = ?", r).First(&roleExist).Error; err != nil {
			// Jika role belum ada, buat baru
			// Khusus Admin, beri permission user:manage
			perms := []models.Permission{}
			if r == "Admin" {
				// Pastikan permission ada dulu
				perm := models.Permission{Name: "user:manage", Resource: "user", Action: "manage"}
				db.FirstOrCreate(&perm, models.Permission{Name: "user:manage"})
				perms = append(perms, perm)
			}
			
			newRole := models.Role{Name: r, Permissions: perms}
			db.Create(&newRole)
			fmt.Printf("[SEED] Role created: %s\n", r)
		}
	}

	// 2. Seed Admin User
	var adminRole models.Role
	db.Where("name = ?", "Admin").First(&adminRole)

	var userExist models.User
	if err := db.Where("username = ?", "admin").First(&userExist).Error; err != nil {
		// Jika user admin belum ada, buat baru
		hash, _ := helper.HashPassword("admin123") // Password default
		admin := models.User{
			Username:     "admin",
			Email:        "admin@gmail.com",
			PasswordHash: hash,
			FullName:     "Super Admin",
			RoleID:       adminRole.ID,
			IsActive:     true,
		}
		db.Create(&admin)
		fmt.Println("[SEED] User 'admin' created with password 'admin123'")
	}
}

// @title Sistem Pelaporan Prestasi API
// @version 1.0
// @description API Documentation for GOUAS Project
// @host localhost:3000
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Config & Logger
	config.LoadEnv()
	config.InitLogger()

	// 2. Database Connection
	database.ConnectPostgres()
	database.ConnectMongo()
	
	// Migration
	database.Migrate()

	// 3. Dependency Injection (Repository)
	db := database.DB
	mongoDB := database.MongoDB

	// --- JALANKAN SEEDER DI SINI ---
	seedDatabase(db)
	// -------------------------------

	authRepo := repository.NewAuthRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	achievementRepo := repository.NewAchievementRepository(db, mongoDB)
	studentRepo := repository.NewStudentRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	reportRepo := repository.NewReportRepository(db, mongoDB)

	// 4. Dependency Injection (Service)
	authSvc := service.NewAuthService(authRepo)
	adminSvc := service.NewAdminService(adminRepo)
	achievementSvc := service.NewAchievementService(achievementRepo)
	studentSvc := service.NewStudentService(studentRepo)
	lecturerSvc := service.NewLecturerService(lecturerRepo)
	reportSvc := service.NewReportService(reportRepo)

	// 5. Init Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Sistem Pelaporan Prestasi v1.0",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	// 6. SWAGGER ROUTE
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 7. Route Initialization
	route.InitRoutes(app, authSvc, adminSvc, achievementSvc, studentSvc, lecturerSvc, reportSvc)

	// 8. Server Start
	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Swagger UI is available at http://localhost:%s/swagger/index.html", port)
	log.Fatal(app.Listen(":" + port))
}