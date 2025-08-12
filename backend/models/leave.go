package models

import (
	"time"

	"gorm.io/gorm"
)

type LeaveStatus string

const (
	LeavePending  LeaveStatus = "pending"
	LeaveApproved LeaveStatus = "approved"
	LeaveRejected LeaveStatus = "rejected"
)

type LeaveType string

const (
	LeaveTypeVacation LeaveType = "vacation"
	LeaveTypeSick     LeaveType = "sick"
	LeaveTypePersonal LeaveType = "personal"
	LeaveTypeMaternal LeaveType = "maternal"
	LeaveTypePaternal LeaveType = "paternal"
)

type Leave struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Type        LeaveType      `json:"type" gorm:"not null"`
	StartDate   time.Time      `json:"start_date" gorm:"not null"`
	EndDate     time.Time      `json:"end_date" gorm:"not null"`
	Days        int            `json:"days" gorm:"not null"`
	Reason      string         `json:"reason"`
	Status      LeaveStatus    `json:"status" gorm:"default:pending"`
	ApprovedBy  *uint          `json:"approved_by"`
	ApprovedAt  *time.Time     `json:"approved_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	User     User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Approver *User `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
}