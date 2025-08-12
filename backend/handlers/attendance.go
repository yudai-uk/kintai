package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yudai-uk/backend/models"
	"gorm.io/gorm"
)

type AttendanceHandler struct {
	db *gorm.DB
}

func NewAttendanceHandler(db *gorm.DB) *AttendanceHandler {
	return &AttendanceHandler{db: db}
}

type ClockInRequest struct {
	ClockIn *time.Time `json:"clock_in"`
	Note    string     `json:"note"`
}

type ClockOutRequest struct {
	ClockOut  *time.Time `json:"clock_out"`
	BreakTime int        `json:"break_time"`
	Note      string     `json:"note"`
}

func (h *AttendanceHandler) ClockIn(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	
	var req ClockInRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	now := time.Now()
	if req.ClockIn == nil {
		req.ClockIn = &now
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var existingAttendance models.Attendance
	if err := h.db.Where("user_id = ? AND date = ?", userID, today).First(&existingAttendance).Error; err == nil {
		if existingAttendance.ClockIn != nil {
			return echo.NewHTTPError(http.StatusConflict, "Already clocked in today")
		}
		existingAttendance.ClockIn = req.ClockIn
		existingAttendance.Note = req.Note
		if err := h.db.Save(&existingAttendance).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update attendance")
		}
		return c.JSON(http.StatusOK, existingAttendance)
	}

	attendance := models.Attendance{
		UserID:  userID,
		Date:    today,
		ClockIn: req.ClockIn,
		Note:    req.Note,
	}

	if err := h.db.Create(&attendance).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create attendance record")
	}

	return c.JSON(http.StatusCreated, attendance)
}

func (h *AttendanceHandler) ClockOut(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	
	var req ClockOutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	now := time.Now()
	if req.ClockOut == nil {
		req.ClockOut = &now
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var attendance models.Attendance
	if err := h.db.Where("user_id = ? AND date = ?", userID, today).First(&attendance).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "No attendance record found for today")
	}

	if attendance.ClockIn == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Must clock in before clocking out")
	}

	if attendance.ClockOut != nil {
		return echo.NewHTTPError(http.StatusConflict, "Already clocked out today")
	}

	attendance.ClockOut = req.ClockOut
	attendance.BreakTime = req.BreakTime
	if req.Note != "" {
		attendance.Note = req.Note
	}

	if err := h.db.Save(&attendance).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update attendance")
	}

	return c.JSON(http.StatusOK, attendance)
}

func (h *AttendanceHandler) GetMyAttendance(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var attendances []models.Attendance
	query := h.db.Where("user_id = ?", userID).Order("date DESC").Offset(offset).Limit(limit)

	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")
	
	if startDate != "" && endDate != "" {
		query = query.Where("date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Find(&attendances).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve attendance records")
	}

	var total int64
	countQuery := h.db.Model(&models.Attendance{}).Where("user_id = ?", userID)
	if startDate != "" && endDate != "" {
		countQuery = countQuery.Where("date BETWEEN ? AND ?", startDate, endDate)
	}
	countQuery.Count(&total)

	response := map[string]interface{}{
		"data":     attendances,
		"page":     page,
		"limit":    limit,
		"total":    total,
		"has_next": int64(page*limit) < total,
	}

	return c.JSON(http.StatusOK, response)
}