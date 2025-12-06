package main

import (
	"gouas/app/model" // Pastikan import model ada
	"gouas/database"
	"gouas/route"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: File .env tidak ditemukan, menggunakan environment system.")
	}

	// 2. Setup Database Connections
	database.ConnectPostgres()
	database.ConnectMongo()

	// 3. Auto Migration (Agar struct Go sinkron dengan tabel DB)
	// Ini penting agar kolom deleted_at terbaca oleh GORM
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
	log.Println("✅ Migrasi Database Berhasil!")

	// 4. Seeding Data (PENTING: Ini yang membuat user Admin)
	database.SeedDatabase()

	// 5. Init Fiber App
	app := fiber.New()

	// 6. Middlewares Dasar
	app.Use(cors.New())   // Agar bisa diakses frontend
	app.Use(logger.New()) // Log setiap request masuk

	// 7. Setup Routes
	route.SetupRoutes(app)

	// 8. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server running on port " + port)
	log.Fatal(app.Listen(":" + port))
}