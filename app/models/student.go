package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	// ID (UUID) -> Ini Primary Key tabel students.
	// Ini yang akan dipanggil oleh tabel 'achievement_references' sebagai foreign key.
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// Relasi ke User
	UserID uuid.UUID `gorm:"type:uuid;not null"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// NIM (STRING) -> Ini data teks biasa (contoh: "NIM-zaky-123")
	// Kita kunci tipe datanya menjadi varchar(100) agar GORM tidak error lagi.
	StudentID string `gorm:"column:student_id;type:varchar(100);unique;not null"`

	ProgramStudy string `gorm:"type:varchar(100)"`
	AcademicYear string `gorm:"type:varchar(10)"`

	// Advisor (Boleh Null)
	AdvisorID *uuid.UUID `gorm:"type:uuid"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
