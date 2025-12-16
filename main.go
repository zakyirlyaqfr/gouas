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
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

// Fungsi untuk seeding data awal (Role, Permission & Admin)
func seedDatabase(db *gorm.DB) {
	// Definisi Role beserta Permission-nya sesuai SRS
	rolePermissions := map[string][]string{
		"Admin": {
			"user:manage",
		},
		"Mahasiswa": {
			"achievement:create",
			"achievement:read",
			"achievement:update",
			"achievement:delete",
		},
		"Dosen Wali": {
			"achievement:read",
			"achievement:verify",
		},
	}

	for roleName, permNames := range rolePermissions {
		// 1. Buat Permission jika belum ada
		var permissions []models.Permission
		for _, permName := range permNames {
			var perm models.Permission
			// Cek apakah permission sudah ada
			if err := db.Where("name = ?", permName).First(&perm).Error; err != nil {
				perm = models.Permission{
					Name:        permName,
					Resource:    permName,
					Action:      "access",
					Description: "Auto generated",
				}
				db.Create(&perm)
			}
			permissions = append(permissions, perm)
		}

		// 2. Buat Role dan Assign Permission
		var role models.Role
		if err := db.Where("name = ?", roleName).Preload("Permissions").First(&role).Error; err != nil {
			newRole := models.Role{Name: roleName, Permissions: permissions}
			db.Create(&newRole)
			fmt.Printf("[SEED] Role created: %s with permissions %v\n", roleName, permNames)
		} else {
			// Update permissions jika sudah ada
			db.Model(&role).Association("Permissions").Replace(permissions)
			fmt.Printf("[SEED] Role updated: %s permissions refreshed\n", roleName)
		}
	}

	// 3. Seed Admin User
	var adminRole models.Role
	db.Where("name = ?", "Admin").First(&adminRole)

	var userExist models.User
	if err := db.Where("username = ?", "admin").First(&userExist).Error; err != nil {
		hash, _ := helper.HashPassword("admin123")
		admin := models.User{
			Username:     "admin",
			Email:        "admin@gmail.com",
			PasswordHash: hash,
			FullName:     "Super Admin",
			RoleID:       adminRole.ID,
			IsActive:     true,
		}
		db.Create(&admin)
		fmt.Println("[SEED] User 'admin' created")
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

	// [UBAH] Inject studentRepo ke achievementService
	achievementSvc := service.NewAchievementService(achievementRepo, studentRepo)

	studentSvc := service.NewStudentService(studentRepo)
	lecturerSvc := service.NewLecturerService(lecturerRepo)
	reportSvc := service.NewReportService(reportRepo)

	// 5. Init Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Sistem Pelaporan Prestasi v1.0",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	// --- [BARU] CONFIG STATIC FILES ---
	// 1. Pastikan folder uploads ada
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}
	// 2. Buka akses URL /uploads agar mengarah ke folder ./uploads
	app.Static("/uploads", "./uploads")
	// ----------------------------------

	// 6. SWAGGER ROUTE
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 7. Route Initialization
	route.InitRoutes(app, authSvc, adminSvc, achievementSvc, studentSvc, lecturerSvc, reportSvc)

	// 8. Server Start
	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Swagger UI is available at http://localhost:%s/swagger/index.html", port)
	log.Fatal(app.Listen(":" + port))
}