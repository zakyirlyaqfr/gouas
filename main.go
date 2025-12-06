package main

import (
	"gouas/database"
	"gouas/helper"
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
	app.Use(cors.New())   // Agar bisa diakses frontend
	app.Use(logger.New()) // Log setiap request masuk

	// 5. Test Route (Untuk memastikan server jalan)
	app.Get("/", func(c *fiber.Ctx) error {
		return helper.SuccessResponse(c, "Server Back-end SRS Mahasiswa Berjalan!", nil)
	})

	// 6. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}