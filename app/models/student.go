package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;unique;not null"`
	User         User      `gorm:"foreignKey:UserID"`
	StudentID    string    `gorm:"type:varchar(20);unique;not null"`
	ProgramStudy string    `gorm:"type:varchar(100)"`
	AcademicYear string    `gorm:"type:varchar(10)"`
	AdvisorID    *uuid.UUID `gorm:"type:uuid"`
	Advisor      *Lecturer  `gorm:"foreignKey:AdvisorID"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}