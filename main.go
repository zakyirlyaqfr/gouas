package main

import (
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

	// 3. Init Fiber App
	app := fiber.New()

	// 4. Middlewares Dasar
	app.Use(cors.New())   // Agar bisa diakses frontend/client lain
	app.Use(logger.New()) // Log setiap request masuk

	// 5. Setup Routes (Panggil fungsi dari package route)
	route.SetupRoutes(app)

	// 6. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server running on port " + port)
	log.Fatal(app.Listen(":" + port))
}