package models

import (
	"time"

	"gorm.io/gorm"
)

type Schedule struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	Date         time.Time      `json:"date" gorm:"not null;index"`
	StartTime    time.Time      `json:"start_time" gorm:"not null"`
	EndTime      time.Time      `json:"end_time" gorm:"not null"`
	BreakTime    int            `json:"break_time" gorm:"default:60"` // minutes
	IsFlexTime   bool           `json:"is_flex_time" gorm:"default:false"`
	Note         string         `json:"note"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (s *Schedule) PlannedHours() float64 {
	duration := s.EndTime.Sub(s.StartTime)
	hours := duration.Hours() - float64(s.BreakTime)/60.0
	if hours < 0 {
		return 0
	}
	return hours
}