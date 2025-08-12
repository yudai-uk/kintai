package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yudai-uk/backend/models"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

type MonthlyReportData struct {
	User              models.User `json:"user"`
	TotalWorkingDays  int         `json:"total_working_days"`
	ActualWorkingDays int         `json:"actual_working_days"`
	TotalWorkingHours string      `json:"total_working_hours"`
	PlannedHours      string      `json:"planned_hours"`
	Overtime          string      `json:"overtime"`
	LeaveDays         int         `json:"leave_days"`
	PendingLeaves     int         `json:"pending_leaves"`
	AttendanceRate    string      `json:"attendance_rate"`
}

type MonthlyReportSummary struct {
	Month            string              `json:"month"`
	TotalEmployees   int                 `json:"total_employees"`
	AverageWorking   string              `json:"average_working_hours"`
	TotalLeaves      int                 `json:"total_leaves"`
	PendingLeaves    int                 `json:"pending_leaves"`
	AverageAttendance string             `json:"average_attendance_rate"`
	Reports          []MonthlyReportData `json:"reports"`
}

func (h *AdminHandler) GetMonthlyReports(c echo.Context) error {
	userRole := c.Get("user_role").(string)
	if userRole != "admin" && userRole != "manager" {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
	}

	month := c.QueryParam("month")
	if month == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Month parameter is required (format: YYYY-MM)")
	}

	parsedTime, err := time.Parse("2006-01", month)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid month format. Expected YYYY-MM")
	}

	startOfMonth := time.Date(parsedTime.Year(), parsedTime.Month(), 1, 0, 0, 0, 0, parsedTime.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	var users []models.User
	if err := h.db.Where("role != ?", "admin").Find(&users).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve users")
	}

	var reports []MonthlyReportData
	totalWorkingHoursSum := 0.0
	totalLeavesSum := 0
	totalPendingLeavesSum := 0
	totalAttendanceRateSum := 0.0

	for _, user := range users {
		reportData := h.generateUserMonthlyReport(user, startOfMonth, endOfMonth)
		reports = append(reports, reportData)

		if hours, err := parseFloat(reportData.TotalWorkingHours); err == nil {
			totalWorkingHoursSum += hours
		}
		totalLeavesSum += reportData.LeaveDays
		totalPendingLeavesSum += reportData.PendingLeaves
		if rate, err := parseFloat(reportData.AttendanceRate); err == nil {
			totalAttendanceRateSum += rate
		}
	}

	summary := MonthlyReportSummary{
		Month:          month,
		TotalEmployees: len(users),
		Reports:        reports,
		TotalLeaves:    totalLeavesSum,
		PendingLeaves:  totalPendingLeavesSum,
	}

	if len(users) > 0 {
		summary.AverageWorking = fmt.Sprintf("%.2f", totalWorkingHoursSum/float64(len(users)))
		summary.AverageAttendance = fmt.Sprintf("%.2f", totalAttendanceRateSum/float64(len(users)))
	} else {
		summary.AverageWorking = "0.00"
		summary.AverageAttendance = "0.00"
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *AdminHandler) generateUserMonthlyReport(user models.User, startOfMonth, endOfMonth time.Time) MonthlyReportData {
	var attendances []models.Attendance
	h.db.Where("user_id = ? AND date BETWEEN ? AND ?", user.ID, startOfMonth, endOfMonth).Find(&attendances)

	var schedules []models.Schedule
	h.db.Where("user_id = ? AND date BETWEEN ? AND ?", user.ID, startOfMonth, endOfMonth).Find(&schedules)

	var leaves []models.Leave
	h.db.Where("user_id = ? AND start_date <= ? AND end_date >= ? AND status = ?", 
		user.ID, endOfMonth, startOfMonth, models.LeaveApproved).Find(&leaves)

	var pendingLeaves []models.Leave
	h.db.Where("user_id = ? AND start_date <= ? AND end_date >= ? AND status = ?", 
		user.ID, endOfMonth, startOfMonth, models.LeavePending).Find(&pendingLeaves)

	totalWorkingDays := len(schedules)
	actualWorkingDays := 0
	totalWorkingHours := 0.0
	plannedHours := 0.0

	for _, attendance := range attendances {
		if attendance.ClockIn != nil && attendance.ClockOut != nil {
			actualWorkingDays++
			totalWorkingHours += attendance.WorkingHours()
		}
	}

	for _, schedule := range schedules {
		plannedHours += schedule.PlannedHours()
	}

	overtime := totalWorkingHours - plannedHours
	if overtime < 0 {
		overtime = 0
	}

	leaveDays := h.calculateLeaveDaysInMonth(leaves, startOfMonth, endOfMonth)

	attendanceRate := 0.0
	if totalWorkingDays > 0 {
		attendanceRate = float64(actualWorkingDays) / float64(totalWorkingDays) * 100
	}

	return MonthlyReportData{
		User:              user,
		TotalWorkingDays:  totalWorkingDays,
		ActualWorkingDays: actualWorkingDays,
		TotalWorkingHours: fmt.Sprintf("%.2f", totalWorkingHours),
		PlannedHours:      fmt.Sprintf("%.2f", plannedHours),
		Overtime:          fmt.Sprintf("%.2f", overtime),
		LeaveDays:         leaveDays,
		PendingLeaves:     len(pendingLeaves),
		AttendanceRate:    fmt.Sprintf("%.2f", attendanceRate),
	}
}

func (h *AdminHandler) calculateLeaveDaysInMonth(leaves []models.Leave, startOfMonth, endOfMonth time.Time) int {
	totalDays := 0
	for _, leave := range leaves {
		start := leave.StartDate
		end := leave.EndDate

		if start.Before(startOfMonth) {
			start = startOfMonth
		}
		if end.After(endOfMonth) {
			end = endOfMonth
		}

		if start.Before(end) || start.Equal(end) {
			days := int(end.Sub(start).Hours()/24) + 1
			totalDays += days
		}
	}
	return totalDays
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}