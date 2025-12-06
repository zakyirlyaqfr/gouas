package model

import "github.com/google/uuid"

type Lecturer struct {
	Base
	UserID     uuid.UUID
	User       User   `gorm:"foreignKey:UserID"`
	LecturerID string `gorm:"type:varchar(20);unique;not null"` // NIP
	Department string `gorm:"type:varchar(100)"`
}

type Student struct {
	Base
	UserID       uuid.UUID
	User         User      `gorm:"foreignKey:UserID"`
	StudentID    string    `gorm:"type:varchar(20);unique;not null"` // NIM
	ProgramStudy string    `gorm:"type:varchar(100)"`
	AcademicYear string    `gorm:"type:varchar(10)"`
	AdvisorID    *uuid.UUID // Pointer karena bisa null di awal
	Advisor      *Lecturer  `gorm:"foreignKey:AdvisorID"`
}