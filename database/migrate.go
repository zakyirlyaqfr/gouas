package database

import (
	"gouas/app/models"
	"log"
)

func Migrate() {
	err := DB.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.User{},
		&models.Student{}, // Dibuat DULUAN
		&models.Lecturer{},
		&models.AchievementReference{}, // Dibuat TERAKHIR
	)

	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	log.Println("Database migration completed successfully")
}
