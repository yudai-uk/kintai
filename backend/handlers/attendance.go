package handlers

import (
    "net/http"
    "strconv"
    "time"
    "strings"

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

type WorkModeRequest struct {
    Mode string `json:"mode"` // "office" or "remote"
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
            // Make clock-in idempotent: return current state as 200 OK
            return c.JSON(http.StatusOK, existingAttendance)
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
    if attendance.BreakStart != nil && attendance.BreakEnd == nil {
        return echo.NewHTTPError(http.StatusBadRequest, "End break before clocking out")
    }
    if attendance.OutStart != nil && attendance.OutEnd == nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Return from out before clocking out")
    }

    attendance.ClockOut = req.ClockOut
    if req.Note != "" {
        attendance.Note = req.Note
    }

	if err := h.db.Save(&attendance).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update attendance")
	}

	return c.JSON(http.StatusOK, attendance)
}

// BreakStart marks the beginning of a break.
func (h *AttendanceHandler) BreakStart(c echo.Context) error {
    userID := c.Get("user_id").(uint)
    now := time.Now()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    var attendance models.Attendance
    if err := h.db.Where("user_id = ? AND date = ?", userID, today).First(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Must clock in before starting break")
    }
    if attendance.ClockOut != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Already clocked out")
    }
    if attendance.BreakStart != nil {
        return echo.NewHTTPError(http.StatusConflict, "Break already started")
    }
    if attendance.BreakEnd != nil { // limit to one break per day
        return echo.NewHTTPError(http.StatusBadRequest, "Break already completed today")
    }
    if attendance.OutStart != nil && attendance.OutEnd == nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Cannot start break while out")
    }
    if attendance.ClockIn == nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Must clock in first")
    }
    attendance.BreakStart = &now
    if err := h.db.Save(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start break")
    }
    return c.JSON(http.StatusOK, attendance)
}

// BreakEnd ends break and accumulates break minutes.
func (h *AttendanceHandler) BreakEnd(c echo.Context) error {
    userID := c.Get("user_id").(uint)
    now := time.Now()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    var attendance models.Attendance
    if err := h.db.Where("user_id = ? AND date = ?", userID, today).First(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "No attendance record for today")
    }
    if attendance.BreakStart == nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Break not started")
    }
    if attendance.BreakEnd != nil { // already ended once
        return echo.NewHTTPError(http.StatusConflict, "Break already ended today")
    }
    if attendance.ClockOut != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Already clocked out")
    }
    if now.Before(*attendance.BreakStart) {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid break end time")
    }
    // Keep BreakStart for audit; set BreakEnd (no cumulative tracking)
    attendance.BreakEnd = &now
    if err := h.db.Save(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to end break")
    }
    return c.JSON(http.StatusOK, attendance)
}

// GoOut marks going out.
func (h *AttendanceHandler) GoOut(c echo.Context) error {
    userID := c.Get("user_id").(uint)
    now := time.Now()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    var attendance models.Attendance
    if err := h.db.Where("user_id = ? AND date = ?", userID, today).First(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Must clock in before going out")
    }
    if attendance.ClockOut != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Already clocked out")
    }
    if attendance.OutStart != nil && attendance.OutEnd == nil {
        return echo.NewHTTPError(http.StatusConflict, "Already out")
    }
    if attendance.OutEnd != nil { // limit to one outing per day
        return echo.NewHTTPError(http.StatusBadRequest, "Outing already completed today")
    }
    if attendance.BreakStart != nil && attendance.BreakEnd == nil {
        return echo.NewHTTPError(http.StatusBadRequest, "End break before going out")
    }
    attendance.OutStart = &now
    if err := h.db.Save(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to mark out")
    }
    return c.JSON(http.StatusOK, attendance)
}

// ReturnFromOut marks returning from going out.
func (h *AttendanceHandler) ReturnFromOut(c echo.Context) error {
    userID := c.Get("user_id").(uint)
    now := time.Now()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    var attendance models.Attendance
    if err := h.db.Where("user_id = ? AND date = ?", userID, today).First(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "No attendance record for today")
    }
    if attendance.OutStart == nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Not currently out")
    }
    if attendance.OutEnd != nil {
        return echo.NewHTTPError(http.StatusConflict, "Already returned today")
    }
    if attendance.ClockOut != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Already clocked out")
    }
    if now.Before(*attendance.OutStart) {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid return time")
    }
    // Keep OutStart for audit; set OutEnd
    attendance.OutEnd = &now
    if err := h.db.Save(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to return")
    }
    return c.JSON(http.StatusOK, attendance)
}

// SetWorkMode toggles or sets work mode for the day (office/remote).
func (h *AttendanceHandler) SetWorkMode(c echo.Context) error {
    userID := c.Get("user_id").(uint)
    var req WorkModeRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
    }
    mode := strings.ToLower(strings.TrimSpace(req.Mode))
    if mode != "office" && mode != "remote" {
        return echo.NewHTTPError(http.StatusBadRequest, "mode must be 'office' or 'remote'")
    }

    now := time.Now()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    var attendance models.Attendance
    if err := h.db.Where("user_id = ? AND date = ?", userID, today).First(&attendance).Error; err != nil {
        // If not exists yet (no clock in), create a record for the day with chosen mode
        attendance = models.Attendance{UserID: userID, Date: today, WorkMode: mode}
        if err := h.db.Create(&attendance).Error; err != nil {
            return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set work mode")
        }
        return c.JSON(http.StatusOK, attendance)
    }
    attendance.WorkMode = mode
    if err := h.db.Save(&attendance).Error; err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update work mode")
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
