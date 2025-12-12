package database

import (
	"gouas/app/models"
	"log"
)

func Migrate() {
	// AutoMigrate hanya membuat tabel/kolom yang belum ada
	err := DB.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.User{},
		&models.Lecturer{},
		&models.Student{},
		&models.AchievementReference{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed successfully")
}