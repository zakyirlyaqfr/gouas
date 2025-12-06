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
	StudentID          uuid.UUID         `gorm:"type:uuid" json:"student_id"`
	Student            Student           `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	MongoAchievementID string            `gorm:"type:varchar(24);not null" json:"mongo_achievement_id"`
	Status             AchievementStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	SubmittedAt        *time.Time        `json:"submitted_at"`
	VerifiedAt         *time.Time        `json:"verified_at"`
	VerifiedBy         *uuid.UUID        `gorm:"type:uuid" json:"verified_by"`
	Verifier           *User             `gorm:"foreignKey:VerifiedBy" json:"verifier,omitempty"`
	RejectionNote      string            `json:"rejection_note"`
}