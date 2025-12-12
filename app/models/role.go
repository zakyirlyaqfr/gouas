package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string       `gorm:"type:varchar(50);unique;not null"`
	Description string       `gorm:"type:text"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}