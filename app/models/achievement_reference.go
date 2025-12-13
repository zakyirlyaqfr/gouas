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
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// Foreign Key (Menyimpan UUID dari Student.ID)
	StudentID uuid.UUID `gorm:"type:uuid;not null"`

	// Relasi (GORM akan mencocokkan StudentID diatas dengan Student.ID)
	Student Student `gorm:"foreignKey:StudentID;references:ID"`

	MongoAchievementID string            `gorm:"type:varchar(50);not null"`
	Status             AchievementStatus `gorm:"type:varchar(20);default:'draft'"`

	SubmittedAt *time.Time
	VerifiedAt  *time.Time

	VerifiedBy *uuid.UUID `gorm:"type:uuid"`
	Verifier   *User      `gorm:"foreignKey:VerifiedBy"`

	RejectionNote string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
