package model

// Base struct sudah menghandle ID (UUID), jadi file ini tidak butuh import uuid secara eksplisit
// kecuali Anda menambahkan field baru bertipe uuid.UUID

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