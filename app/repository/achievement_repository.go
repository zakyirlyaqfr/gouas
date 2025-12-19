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
	FindAllReferences() ([]models.AchievementReference, error)
	FindReferencesByStudentID(studentID uuid.UUID) ([]models.AchievementReference, error)
	// [BARU] Mencari berdasarkan list Student ID (untuk Dosen Wali)
	FindReferencesByStudentIDs(studentIDs []uuid.UUID) ([]models.AchievementReference, error)
	GetMongoDetail(mongoID string) (*models.Achievement, error)
	UpdateMongo(mongoID string, data models.Achievement) error
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

	achievement.ID = primitive.NewObjectID()
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()
	if achievement.Attachments == nil {
		achievement.Attachments = []models.Attachment{}
	}
	achievement.StudentID = studentID.String()

	_, err := r.mongo.InsertOne(ctx, achievement)
	if err != nil {
		return nil, err
	}

	ref := models.AchievementReference{
		StudentID:          studentID,
		MongoAchievementID: achievement.ID.Hex(),
		Status:             models.StatusDraft,
	}

	err = r.pg.Create(&ref).Error
	if err != nil {
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

func (r *achievementRepository) FindAllReferences() ([]models.AchievementReference, error) {
	var refs []models.AchievementReference
	err := r.pg.Preload("Student.User").Order("created_at desc").Find(&refs).Error
	return refs, err
}

func (r *achievementRepository) FindReferencesByStudentID(studentID uuid.UUID) ([]models.AchievementReference, error) {
	var refs []models.AchievementReference
	err := r.pg.Where("student_id = ?", studentID).Order("created_at desc").Find(&refs).Error
	return refs, err
}

func (r *achievementRepository) FindReferencesByStudentIDs(studentIDs []uuid.UUID) ([]models.AchievementReference, error) {
	var refs []models.AchievementReference
	if len(studentIDs) == 0 {
		return refs, nil
	}
	err := r.pg.Preload("Student.User").Where("student_id IN ?", studentIDs).Order("created_at desc").Find(&refs).Error
	return refs, err
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
	objID, _ := primitive.ObjectIDFromHex(mongoIDHex)
	update := bson.M{
		"$push": bson.M{"attachments": attachment},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	_, err := r.mongo.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *achievementRepository) SoftDelete(id uuid.UUID) error {
	return r.pg.Model(&models.AchievementReference{}).Where("id = ?", id).Update("status", models.StatusDeleted).Error
}

func (r *achievementRepository) GetMongoDetail(mongoID string) (*models.Achievement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	objID, _ := primitive.ObjectIDFromHex(mongoID)
	var achievement models.Achievement
	err := r.mongo.FindOne(ctx, bson.M{"_id": objID}).Decode(&achievement)
	return &achievement, err
}

func (r *achievementRepository) UpdateMongo(mongoID string, data models.Achievement) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	objID, _ := primitive.ObjectIDFromHex(mongoID)
	update := bson.M{
		"$set": bson.M{
			"title":           data.Title,
			"description":     data.Description,
			"achievementType": data.AchievementType,
			"details":         data.Details,
			"tags":            data.Tags,
			"updatedAt":       time.Now(),
		},
	}
	_, err := r.mongo.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}