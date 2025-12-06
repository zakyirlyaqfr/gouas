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
	FindReferenceByID(id uuid.UUID) (*model.AchievementReference, error) // New
	FindReferenceByMongoID(mongoID string) (*model.AchievementReference, error)
	FindReferencesByStudentID(studentID uuid.UUID) ([]model.AchievementReference, error)
	
	// Dosen Operations (New)
	FindReferencesByAdvisorID(advisorID uuid.UUID) ([]model.AchievementReference, error)
	VerifyAchievement(id uuid.UUID, verifierID uuid.UUID) error
	RejectAchievement(id uuid.UUID, note string) error
	
	UpdateReferenceStatus(id uuid.UUID, status model.AchievementStatus) error
	SoftDeleteReference(id uuid.UUID) error
	
	// Helper Profile
	FindStudentByUserID(userID uuid.UUID) (*model.Student, error)
	FindLecturerByUserID(userID uuid.UUID) (*model.Lecturer, error) // New
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

// ... (Existing Mongo Methods tetap sama) ...
func (r *achievementRepository) CreateMongo(data *model.MongoAchievement) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data.CreatedAt = time.Now(); data.UpdatedAt = time.Now()
	res, err := r.mongo.InsertOne(ctx, data)
	if err != nil { return "", err }
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

// ... (Existing Postgres Methods) ...

func (r *achievementRepository) CreateReference(ref *model.AchievementReference) error {
	return r.pg.Create(ref).Error
}

func (r *achievementRepository) FindReferenceByID(id uuid.UUID) (*model.AchievementReference, error) {
	var ref model.AchievementReference
	err := r.pg.Preload("Student").First(&ref, "id = ?", id).Error
	return &ref, err
}

func (r *achievementRepository) FindReferenceByMongoID(mongoID string) (*model.AchievementReference, error) {
	var ref model.AchievementReference
	err := r.pg.Where("mongo_achievement_id = ?", mongoID).First(&ref).Error
	return &ref, err
}

func (r *achievementRepository) FindReferencesByStudentID(studentID uuid.UUID) ([]model.AchievementReference, error) {
	var refs []model.AchievementReference
	err := r.pg.Where("student_id = ? AND status != ?", studentID, model.StatusDeleted).Find(&refs).Error
	return refs, err
}

func (r *achievementRepository) UpdateReferenceStatus(id uuid.UUID, status model.AchievementStatus) error {
	return r.pg.Model(&model.AchievementReference{}).Where("id = ?", id).Update("status", status).Error
}

func (r *achievementRepository) SoftDeleteReference(id uuid.UUID) error {
	return r.pg.Model(&model.AchievementReference{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": model.StatusDeleted, "deleted_at": time.Now()}).Error
}

func (r *achievementRepository) FindStudentByUserID(userID uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.pg.Where("user_id = ?", userID).First(&student).Error
	return &student, err
}

// --- NEW METHODS FOR TAHAP 6 ---

func (r *achievementRepository) FindLecturerByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.pg.Where("user_id = ?", userID).First(&lecturer).Error
	return &lecturer, err
}

func (r *achievementRepository) FindReferencesByAdvisorID(advisorID uuid.UUID) ([]model.AchievementReference, error) {
	var refs []model.AchievementReference
	// Join dengan tabel Students untuk filter by advisor_id
	// Hanya ambil yang statusnya 'submitted', 'verified', 'rejected' (Draft tidak perlu dilihat dosen)
	err := r.pg.Joins("JOIN students ON students.id = achievement_references.student_id").
		Where("students.advisor_id = ? AND achievement_references.status IN (?, ?, ?)", 
			advisorID, model.StatusSubmitted, model.StatusVerified, model.StatusRejected).
		Preload("Student"). // Load data mahasiswa
		Find(&refs).Error
	return refs, err
}

func (r *achievementRepository) VerifyAchievement(id uuid.UUID, verifierID uuid.UUID) error {
	return r.pg.Model(&model.AchievementReference{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.StatusVerified,
			"verified_by": verifierID,
			"verified_at": time.Now(),
		}).Error
}

func (r *achievementRepository) RejectAchievement(id uuid.UUID, note string) error {
	return r.pg.Model(&model.AchievementReference{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         model.StatusRejected,
			"rejection_note": note,
		}).Error
}