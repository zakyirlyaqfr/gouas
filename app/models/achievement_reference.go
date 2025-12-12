package models

import (
	"time"

	"github.com/google/uuid"
)

type AchievementStatus string

const (
	StatusDraft     AchievementStatus = "draft"
	StatusSubmitted AchievementStatus = "submitted"
	StatusVerified  AchievementStatus = "verified"
	StatusRejected  AchievementStatus = "rejected"
	StatusDeleted   AchievementStatus = "deleted"
)

type AchievementReference struct {
	ID                 uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	
	// Relasi ke Student (UUID)
	StudentID          uuid.UUID         `gorm:"type:uuid;not null"`
	// KITA KOMENTAR DULU AGAR MIGRASI LANCAR (Hapus komen ini nanti jika tabel sudah jadi)
	// Student            Student           `gorm:"foreignKey:StudentID;references:ID"` 
	
	MongoAchievementID string            `gorm:"type:varchar(50);not null"`
	Status             AchievementStatus `gorm:"type:varchar(20);default:'draft'"`
	
	SubmittedAt        *time.Time
	VerifiedAt         *time.Time
	
	VerifiedBy         *uuid.UUID        `gorm:"type:uuid"`
	Verifier           *User             `gorm:"foreignKey:VerifiedBy"`
	
	RejectionNote      string            `gorm:"type:text"`
	
	CreatedAt          time.Time
	UpdatedAt          time.Time
}