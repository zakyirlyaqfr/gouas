package repository

import (
	"context"
	"gouas/app/models"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	Create(achievement models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error)
	FindReferenceByID(id uuid.UUID) (*models.AchievementReference, error)
	UpdateStatus(id uuid.UUID, status models.AchievementStatus) error
	Verify(id uuid.UUID, verifierID uuid.UUID) error
	Reject(id uuid.UUID, note string) error
	AddAttachment(mongoID string, attachment models.Attachment) error
	SoftDelete(id uuid.UUID) error
}

type achievementRepository struct {
	pg    *gorm.DB
	mongo *mongo.Collection
}

func NewAchievementRepository(pg *gorm.DB, mongoDB *mongo.Database) AchievementRepository {
	return &achievementRepository{
		pg:    pg,
		mongo: mongoDB.Collection("achievements"),
	}
}

func (r *achievementRepository) Create(achievement models.Achievement, studentID uuid.UUID) (*models.AchievementReference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Insert ke MongoDB
	achievement.ID = primitive.NewObjectID()
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	// Konversi UUID Student ke String untuk referensi di Mongo
	achievement.StudentID = studentID.String()

	_, err := r.mongo.InsertOne(ctx, achievement)
	if err != nil {
		return nil, err
	}

	// 2. Insert ke PostgreSQL (Reference)
	ref := models.AchievementReference{
		StudentID:          studentID,
		MongoAchievementID: achievement.ID.Hex(),
		Status:             models.StatusDraft,
	}

	err = r.pg.Create(&ref).Error
	if err != nil {
		// Manual Rollback: Hapus data di Mongo jika PG gagal
		r.mongo.DeleteOne(ctx, bson.M{"_id": achievement.ID})
		return nil, err
	}

	return &ref, nil
}

func (r *achievementRepository) FindReferenceByID(id uuid.UUID) (*models.AchievementReference, error) {
	var ref models.AchievementReference
	err := r.pg.Preload("Student").First(&ref, "id = ?", id).Error
	return &ref, err
}

func (r *achievementRepository) UpdateStatus(id uuid.UUID, status models.AchievementStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == models.StatusSubmitted {
		now := time.Now()
		updates["submitted_at"] = &now
	}

	return r.pg.Model(&models.AchievementReference{}).Where("id = ?", id).Updates(updates).Error
}

func (r *achievementRepository) Verify(id uuid.UUID, verifierID uuid.UUID) error {
	now := time.Now()
	return r.pg.Model(&models.AchievementReference{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      models.StatusVerified,
		"verified_by": verifierID,
		"verified_at": now,
		"updated_at":  now,
	}).Error
}

func (r *achievementRepository) Reject(id uuid.UUID, note string) error {
	return r.pg.Model(&models.AchievementReference{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         models.StatusRejected,
		"rejection_note": note,
		"updated_at":     time.Now(),
	}).Error
}

func (r *achievementRepository) AddAttachment(mongoIDHex string, attachment models.Attachment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(mongoIDHex)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$push": bson.M{"attachments": attachment},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err = r.mongo.UpdateOne(ctx, filter, update)
	return err
}

func (r *achievementRepository) SoftDelete(id uuid.UUID) error {
	return r.pg.Model(&models.AchievementReference{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     models.StatusDeleted,
			"updated_at": time.Now(),
		}).Error
}