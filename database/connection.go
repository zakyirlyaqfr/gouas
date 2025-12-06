package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global Variables untuk akses DB dari mana saja
var DB *gorm.DB
var Mongo *mongo.Database

// ConnectPostgres menghubungkan ke PostgreSQL
func ConnectPostgres() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("❌ Gagal connect ke PostgreSQL: ", err)
	}

	log.Println("✅ Berhasil connect ke PostgreSQL!")
}

// ConnectMongo menghubungkan ke MongoDB
func ConnectMongo() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Gagal membuat client MongoDB: ", err)
	}

	// Ping database untuk memastikan koneksi hidup
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ Gagal ping ke MongoDB: ", err)
	}

	Mongo = client.Database(dbName)
	log.Println("✅ Berhasil connect ke MongoDB!")
}