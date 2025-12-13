package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// PERBAIKAN TOTAL: Ganti nama field jadi NIM (bukan StudentID)
	// Ini agar GORM tidak menganggapnya sebagai Foreign Key
	NIM          string    `gorm:"column:nim;type:varchar(100);unique;not null"`

	ProgramStudy string    `gorm:"type:varchar(100)"`
	AcademicYear string    `gorm:"type:varchar(10)"`
	AdvisorID    *uuid.UUID `gorm:"type:uuid"`
	Advisor      *Lecturer  `gorm:"foreignKey:AdvisorID"`

	CreatedAt    time.Time
	UpdatedAt    time.Time
}