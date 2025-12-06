package model

import "github.com/google/uuid"

type Role struct {
	Base
	Name        string       `gorm:"type:varchar(50);unique;not null"`
	Description string       `gorm:"type:text"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

type Permission struct {
	Base
	Name        string `gorm:"type:varchar(100);unique;not null"`
	Resource    string `gorm:"type:varchar(50);not null"` // e.g., achievement
	Action      string `gorm:"type:varchar(50);not null"` // e.g., create
	Description string `gorm:"type:text"`
}