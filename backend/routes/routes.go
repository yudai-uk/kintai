package routes

import (
    "os"

    "github.com/labstack/echo/v4"
    "github.com/yudai-uk/backend/handlers"
    appmw "github.com/yudai-uk/backend/middleware"
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
    jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
    api.Use(appmw.NewAuthMiddleware(db, jwtSecret))

	api.POST("/attendance", func(c echo.Context) error {
		action := c.QueryParam("action")
		switch action {
		case "clock_out":
			return attendanceHandler.ClockOut(c)
		case "break_start":
			return attendanceHandler.BreakStart(c)
		case "break_end":
			return attendanceHandler.BreakEnd(c)
		case "out":
			return attendanceHandler.GoOut(c)
		case "return":
			return attendanceHandler.ReturnFromOut(c)
		case "workmode":
			return attendanceHandler.SetWorkMode(c)
		default:
			return attendanceHandler.ClockIn(c)
		}
	})
	api.GET("/attendance/me", attendanceHandler.GetMyAttendance)

	api.POST("/leaves", leaveHandler.CreateLeave)
	api.GET("/leaves", leaveHandler.GetLeaves)
	api.PUT("/leaves/:leaveId/status", leaveHandler.UpdateLeaveStatus)

	api.GET("/schedules", scheduleHandler.GetSchedules)

    admin := api.Group("/admin")
    admin.Use(appmw.AdminMiddleware)
	admin.GET("/reports/monthly", adminHandler.GetMonthlyReports)
}
