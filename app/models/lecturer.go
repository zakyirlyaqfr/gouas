package models

import (
	"time"

	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;unique;not null"`
	User       User      `gorm:"foreignKey:UserID"`
	LecturerID string    `gorm:"type:varchar(20);unique;not null"`
	Department string    `gorm:"type:varchar(100)"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}