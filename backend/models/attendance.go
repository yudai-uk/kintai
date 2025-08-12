package models

import (
	"time"

	"gorm.io/gorm"
)

type Attendance struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	Date      time.Time      `json:"date" gorm:"not null;index"`
	ClockIn   *time.Time     `json:"clock_in"`
	ClockOut  *time.Time     `json:"clock_out"`
	BreakTime int            `json:"break_time" gorm:"default:0"` // minutes
	Note      string         `json:"note"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (a *Attendance) WorkingHours() float64 {
	if a.ClockIn == nil || a.ClockOut == nil {
		return 0
	}
	duration := a.ClockOut.Sub(*a.ClockIn)
	hours := duration.Hours() - float64(a.BreakTime)/60.0
	if hours < 0 {
		return 0
	}
	return hours
}