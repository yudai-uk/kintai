package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    SupabaseUID string       `json:"supabase_uid" gorm:"uniqueIndex;size:64"`
    Email     string         `json:"email" gorm:"uniqueIndex;not null"`
    Name      string         `json:"name" gorm:"not null"`
    Role      string         `json:"role" gorm:"default:employee"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Attendances []Attendance `json:"attendances,omitempty" gorm:"foreignKey:UserID"`
	Leaves      []Leave      `json:"leaves,omitempty" gorm:"foreignKey:UserID"`
}
