package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	// ID Utama (UUID)
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	
	// NIM (STRING/VARCHAR)
	StudentID    string    `gorm:"type:varchar(50);unique"` 
	
	ProgramStudy string    `gorm:"type:varchar(100)"`
	AcademicYear string    `gorm:"type:varchar(10)"`
	
	AdvisorID    *uuid.UUID `gorm:"type:uuid"`
	
	CreatedAt    time.Time
	UpdatedAt    time.Time
}