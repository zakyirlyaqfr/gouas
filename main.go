package main

import (
	"gouas/app/model"
	"gouas/database"
	"gouas/route"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger" // swaggo adapter
	"github.com/joho/godotenv"
	
	_ "gouas/docs" // PENTING: Import docs yang digenerate swagger nanti
)

// @title           Sistem Pelaporan Prestasi Mahasiswa API
// @version         1.0
// @description     API Server untuk SRS Prestasi Mahasiswa (Hybrid Database).
// @contact.name    Tim Backend
// @contact.email   support@gouas.com
// @host            localhost:3000
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: File .env tidak ditemukan, menggunakan environment system.")
	}

	// 2. Setup Database Connections
	database.ConnectPostgres()
	database.ConnectMongo()

	// 3. Auto Migration
	errMigrate := database.DB.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.User{},
		&model.Lecturer{},
		&model.Student{},
		&model.AchievementReference{},
	)
	if errMigrate != nil {
		log.Fatal("❌ Gagal Migrasi Database: ", errMigrate)
	}

	// 4. Seeding Data
	database.SeedDatabase()

	// 5. Init Fiber App
	app := fiber.New()

	// 6. Middlewares Dasar
	app.Use(cors.New())
	app.Use(logger.New())

	// 7. Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault) // Akses di http://localhost:3000/swagger/index.html

	// 8. Setup Routes
	route.SetupRoutes(app)

	// 9. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server running on port " + port)
	log.Fatal(app.Listen(":" + port))
}