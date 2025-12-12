package models

import (
	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(100);unique;not null"`
	Resource    string    `gorm:"type:varchar(50);not null"`
	Action      string    `gorm:"type:varchar(50);not null"`
	Description string    `gorm:"type:text"`
}