package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base berisi field standar yang dimiliki hampir semua tabel
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate hook untuk generate UUID otomatis jika belum ada
func (base *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}