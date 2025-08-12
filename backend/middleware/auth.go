package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIDHeader := c.Request().Header.Get("X-User-ID")
		if userIDHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing user ID")
		}

		userID, err := strconv.ParseUint(userIDHeader, 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
		}

		userRole := c.Request().Header.Get("X-User-Role")
		if userRole == "" {
			userRole = "employee"
		}

		c.Set("user_id", uint(userID))
		c.Set("user_role", userRole)

		return next(c)
	}
}

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userRole := c.Get("user_role").(string)
		if userRole != "admin" && userRole != "manager" {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}

		return next(c)
	}
}