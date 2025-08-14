package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/yudai-uk/backend/handlers"
	"github.com/yudai-uk/backend/middleware"
	"gorm.io/gorm"
)

func SetupRoutes(e *echo.Echo, db *gorm.DB) {
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "OK",
			"message": "Backend is running",
		})
	})

	attendanceHandler := handlers.NewAttendanceHandler(db)
	leaveHandler := handlers.NewLeaveHandler(db)
	scheduleHandler := handlers.NewScheduleHandler(db)
	adminHandler := handlers.NewAdminHandler(db)

	api := e.Group("/api/v1")
	api.Use(middleware.AuthMiddleware)

	api.POST("/attendance", func(c echo.Context) error {
		action := c.QueryParam("action")
		if action == "clock_out" {
			return attendanceHandler.ClockOut(c)
		}
		return attendanceHandler.ClockIn(c)
	})
	api.GET("/attendance/me", attendanceHandler.GetMyAttendance)

	api.POST("/leaves", leaveHandler.CreateLeave)
	api.GET("/leaves", leaveHandler.GetLeaves)
	api.PUT("/leaves/:leaveId/status", leaveHandler.UpdateLeaveStatus)

	api.GET("/schedules", scheduleHandler.GetSchedules)

	admin := api.Group("/admin")
	admin.Use(middleware.AdminMiddleware)
	admin.GET("/reports/monthly", adminHandler.GetMonthlyReports)
}