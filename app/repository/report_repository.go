package repository

import (
	"context"
	"gouas/app/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type ReportRepository interface {
	GetAchievementStats() (map[string]interface{}, error)
}

type reportRepository struct {
	pg    *gorm.DB
	mongo *mongo.Collection
}

func NewReportRepository(pg *gorm.DB, mongoDB *mongo.Database) ReportRepository {
	return &reportRepository{
		pg:    pg,
		mongo: mongoDB.Collection("achievements"),
	}
}

func (r *reportRepository) GetAchievementStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 1. PostgreSQL: Count by Status
	var statusCounts []struct {
		Status string
		Count  int
	}
	r.pg.Model(&models.AchievementReference{}).Select("status, count(*) as count").Group("status").Scan(&statusCounts)
	stats["by_status"] = statusCounts

	// 2. MongoDB: Count by Type (Aggregation)
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.mongo.Aggregate(context.Background(), pipeline)
	if err == nil {
		var typeCounts []bson.M
		if err = cursor.All(context.Background(), &typeCounts); err == nil {
			stats["by_type"] = typeCounts
		}
	}

	return stats, nil
}