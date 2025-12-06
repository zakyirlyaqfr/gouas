package model

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
	Base
	StudentID          uuid.UUID         `gorm:"type:uuid"` // <--- Update baris ini
	Student            Student           `gorm:"foreignKey:StudentID"`
	MongoAchievementID string            `gorm:"type:varchar(24);not null"`
	Status             AchievementStatus `gorm:"type:varchar(20);default:'draft'"`
	SubmittedAt        *time.Time
	VerifiedAt         *time.Time
	VerifiedBy         *uuid.UUID        `gorm:"type:uuid"` // <--- Update baris ini
	Verifier           *User             `gorm:"foreignKey:VerifiedBy"`
	RejectionNote      string
}