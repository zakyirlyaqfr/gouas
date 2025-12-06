package model

import "github.com/google/uuid"

type User struct {
	Base
	Username     string `gorm:"type:varchar(50);unique;not null"`
	Email        string `gorm:"type:varchar(100);unique;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	FullName     string `gorm:"type:varchar(100);not null"`
	RoleID       uuid.UUID
	Role         Role   `gorm:"foreignKey:RoleID"`
	IsActive     bool   `gorm:"default:true"`
}