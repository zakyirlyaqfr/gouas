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
		var permissions []models.Permission
		for _, permName := range permNames {
			var perm models.Permission
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

		var role models.Role
		if err := db.Where("name = ?", roleName).Preload("Permissions").First(&role).Error; err != nil {
			newRole := models.Role{Name: roleName, Permissions: permissions}
			db.Create(&newRole)
			fmt.Printf("[SEED] Role created: %s\n", roleName)
		} else {
			db.Model(&role).Association("Permissions").Replace(permissions)
			fmt.Printf("[SEED] Role updated: %s\n", roleName)
		}
	}

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
func main() {
	config.LoadEnv()
	config.InitLogger()
	database.ConnectPostgres()
	database.ConnectMongo()
	database.Migrate()

	db := database.DB
	mongoDB := database.MongoDB

	seedDatabase(db)

	authRepo := repository.NewAuthRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	achievementRepo := repository.NewAchievementRepository(db, mongoDB)
	studentRepo := repository.NewStudentRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	reportRepo := repository.NewReportRepository(db, mongoDB)

	authSvc := service.NewAuthService(authRepo)
	adminSvc := service.NewAdminService(adminRepo)
	
	// Inject studentRepo ke achievementService
	achievementSvc := service.NewAchievementService(achievementRepo, studentRepo)

	// [FIX] Inject achievementRepo ke studentService
	studentSvc := service.NewStudentService(studentRepo, achievementRepo)
	
	lecturerSvc := service.NewLecturerService(lecturerRepo)
	
	// [FIX] Inject achievementRepo ke reportService
	reportSvc := service.NewReportService(reportRepo, achievementRepo)

	app := fiber.New(fiber.Config{
		AppName: "Sistem Pelaporan Prestasi v1.0",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}
	app.Static("/uploads", "./uploads")

	app.Get("/swagger/*", swagger.HandlerDefault)

	route.InitRoutes(app, authSvc, adminSvc, achievementSvc, studentSvc, lecturerSvc, reportSvc)

	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Swagger UI is available at http://localhost:%s/swagger/index.html", port)
	log.Fatal(app.Listen(":" + port))
}