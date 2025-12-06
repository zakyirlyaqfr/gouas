package repository

import (
	"context"
	"gouas/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type ReportRepository interface {
	CountByStatus() (map[string]int, error)
	CountByType() (map[string]int, error)
}

type reportRepository struct {
	pg    *gorm.DB
	mongo *mongo.Collection
}

func NewReportRepository() ReportRepository {
	return &reportRepository{
		pg:    database.DB,
		mongo: database.Mongo.Collection("achievements"),
	}
}

// Hitung berdasarkan Status (PostgreSQL)
func (r *reportRepository) CountByStatus() (map[string]int, error) {
	var results []struct {
		Status string
		Count  int
	}
	
	// Query: SELECT status, COUNT(*) FROM achievement_references WHERE deleted_at IS NULL GROUP BY status
	err := r.pg.Table("achievement_references").
		Select("status, count(*) as count").
		Where("deleted_at IS NULL").
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	for _, row := range results {
		stats[row.Status] = row.Count
	}
	return stats, nil
}

// Hitung berdasarkan Tipe (MongoDB Aggregation)
func (r *reportRepository) CountByType() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pipeline Aggregation
	pipeline := mongo.Pipeline{
		// 1. Filter yang tidak deleted
		{{Key: "$match", Value: bson.D{{Key: "deletedAt", Value: nil}}}},
		// 2. Group by achievementType
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.mongo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats[result.ID] = result.Count
		}
	}

	return stats, nil
}