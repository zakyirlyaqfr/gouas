package repository

import (
	"context"
	"gouas/app/model"
	"gouas/database"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	// MongoDB Operations
	CreateMongo(data *model.MongoAchievement) (string, error)
	FindMongoByID(mongoID string) (*model.MongoAchievement, error)
	UpdateMongo(mongoID string, updateData bson.M) error
	SoftDeleteMongo(mongoID string) error
	
	// PostgreSQL Operations
	CreateReference(ref *model.AchievementReference) error
	FindReferenceByMongoID(mongoID string) (*model.AchievementReference, error)
	FindReferencesByStudentID(studentID uuid.UUID) ([]model.AchievementReference, error)
	UpdateReferenceStatus(id uuid.UUID, status model.AchievementStatus) error
	SoftDeleteReference(id uuid.UUID) error
	
	// Helper Profile
	FindStudentByUserID(userID uuid.UUID) (*model.Student, error)
}

type achievementRepository struct {
	pg    *gorm.DB
	mongo *mongo.Collection
}

func NewAchievementRepository() AchievementRepository {
	return &achievementRepository{
		pg:    database.DB,
		mongo: database.Mongo.Collection("achievements"),
	}
}

// --- MongoDB Impl ---
func (r *achievementRepository) CreateMongo(data *model.MongoAchievement) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	
	res, err := r.mongo.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *achievementRepository) FindMongoByID(mongoID string) (*model.MongoAchievement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(mongoID)
	var result model.MongoAchievement
	err := r.mongo.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
	return &result, err
}

func (r *achievementRepository) UpdateMongo(mongoID string, updateData bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(mongoID)
	updateData["updatedAt"] = time.Now()
	
	_, err := r.mongo.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	return err
}

func (r *achievementRepository) SoftDeleteMongo(mongoID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	objID, _ := primitive.ObjectIDFromHex(mongoID)
	_, err := r.mongo.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"deletedAt": time.Now()}})
	return err
}

// --- PostgreSQL Impl ---
func (r *achievementRepository) CreateReference(ref *model.AchievementReference) error {
	return r.pg.Create(ref).Error
}

func (r *achievementRepository) FindReferenceByMongoID(mongoID string) (*model.AchievementReference, error) {
	var ref model.AchievementReference
	err := r.pg.Where("mongo_achievement_id = ?", mongoID).First(&ref).Error
	return &ref, err
}

func (r *achievementRepository) FindReferencesByStudentID(studentID uuid.UUID) ([]model.AchievementReference, error) {
	var refs []model.AchievementReference
	// Preload Data Student untuk info tambahan jika perlu
	err := r.pg.Where("student_id = ? AND status != ?", studentID, model.StatusDeleted).Find(&refs).Error
	return refs, err
}

func (r *achievementRepository) UpdateReferenceStatus(id uuid.UUID, status model.AchievementStatus) error {
	return r.pg.Model(&model.AchievementReference{}).Where("id = ?", id).Update("status", status).Error
}

func (r *achievementRepository) SoftDeleteReference(id uuid.UUID) error {
	// Update status jadi deleted DAN set deleted_at (GORM Soft Delete)
	return r.pg.Model(&model.AchievementReference{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": model.StatusDeleted,
			"deleted_at": time.Now(),
		}).Error
}

func (r *achievementRepository) FindStudentByUserID(userID uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.pg.Where("user_id = ?", userID).First(&student).Error
	return &student, err
}