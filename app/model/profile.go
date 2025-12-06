package model

import "github.com/google/uuid"

type Lecturer struct {
	Base
	UserID     uuid.UUID `gorm:"type:uuid"`
	User       User      `gorm:"foreignKey:UserID"`
	LecturerID string    `gorm:"type:varchar(20);unique;not null"`
	Department string    `gorm:"type:varchar(100)"`
}

type Student struct {
	Base
	UserID       uuid.UUID  `gorm:"type:uuid"`
	User         User       `gorm:"foreignKey:UserID"`
	
	// UBAH DARI StudentID JADI NIM (Agar GORM tidak bingung)
	NIM          string     `gorm:"type:varchar(20);unique;not null;column:nim"` 
	
	ProgramStudy string     `gorm:"type:varchar(100)"`
	AcademicYear string     `gorm:"type:varchar(10)"`
	AdvisorID    *uuid.UUID `gorm:"type:uuid"`
	Advisor      *Lecturer  `gorm:"foreignKey:AdvisorID"`
}