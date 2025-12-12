package models

import (
	"time"

	"github.com/google/uuid"
)

type Lecturer struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	
	// NIP (STRING/VARCHAR)
	LecturerID   string    `gorm:"type:varchar(50);unique"` 
	
	Department   string    `gorm:"type:varchar(100)"`
	
	CreatedAt    time.Time
	UpdatedAt    time.Time
}