package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yudai-uk/backend/models"
	"gorm.io/gorm"
)

type ScheduleHandler struct {
	db *gorm.DB
}

func NewScheduleHandler(db *gorm.DB) *ScheduleHandler {
	return &ScheduleHandler{db: db}
}

func (h *ScheduleHandler) GetSchedules(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	userRole := c.Get("user_role").(string)

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

	query := h.db.Model(&models.Schedule{}).Preload("User").Where("date BETWEEN ? AND ?", startOfMonth, endOfMonth)

	if userRole == "admin" || userRole == "manager" {
		requestUserID := c.QueryParam("user_id")
		if requestUserID != "" {
			uid, err := strconv.ParseUint(requestUserID, 10, 32)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id parameter")
			}
			query = query.Where("user_id = ?", uid)
		}
	} else {
		query = query.Where("user_id = ?", userID)
	}

	var schedules []models.Schedule
	if err := query.Order("date ASC").Find(&schedules).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve schedules")
	}

	summary := h.calculateScheduleSummary(schedules)

	response := map[string]interface{}{
		"data":    schedules,
		"summary": summary,
		"month":   month,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *ScheduleHandler) calculateScheduleSummary(schedules []models.Schedule) map[string]interface{} {
	totalDays := len(schedules)
	totalPlannedHours := 0.0
	flexTimeDays := 0

	for _, schedule := range schedules {
		totalPlannedHours += schedule.PlannedHours()
		if schedule.IsFlexTime {
			flexTimeDays++
		}
	}

	return map[string]interface{}{
		"total_days":          totalDays,
		"total_planned_hours": fmt.Sprintf("%.2f", totalPlannedHours),
		"flex_time_days":      flexTimeDays,
		"average_hours_per_day": func() string {
			if totalDays > 0 {
				return fmt.Sprintf("%.2f", totalPlannedHours/float64(totalDays))
			}
			return "0.00"
		}(),
	}
}