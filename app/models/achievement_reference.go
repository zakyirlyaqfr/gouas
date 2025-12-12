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
	StudentID          uuid.UUID         `gorm:"type:uuid;not null"`
	Student            Student           `gorm:"foreignKey:StudentID"`
	MongoAchievementID string            `gorm:"type:varchar(24);not null"`
	Status             AchievementStatus `gorm:"type:varchar(20);default:'draft'"`
	SubmittedAt        *time.Time
	VerifiedAt         *time.Time
	VerifiedBy         *uuid.UUID `gorm:"type:uuid"`
	Verifier           *User      `gorm:"foreignKey:VerifiedBy"`
	RejectionNote      string     `gorm:"type:text"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`
}