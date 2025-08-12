package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yudai-uk/backend/models"
	"gorm.io/gorm"
)

type LeaveHandler struct {
	db *gorm.DB
}

func NewLeaveHandler(db *gorm.DB) *LeaveHandler {
	return &LeaveHandler{db: db}
}

type CreateLeaveRequest struct {
	Type      models.LeaveType `json:"type" validate:"required"`
	StartDate time.Time        `json:"start_date" validate:"required"`
	EndDate   time.Time        `json:"end_date" validate:"required"`
	Days      int              `json:"days" validate:"required,min=1"`
	Reason    string           `json:"reason" validate:"required"`
}

type UpdateLeaveStatusRequest struct {
	Status models.LeaveStatus `json:"status" validate:"required"`
}

func (h *LeaveHandler) CreateLeave(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	var req CreateLeaveRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.StartDate.After(req.EndDate) {
		return echo.NewHTTPError(http.StatusBadRequest, "Start date must be before end date")
	}

	if req.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot apply for leave in the past")
	}

	var overlappingLeaves []models.Leave
	if err := h.db.Where("user_id = ? AND status != ? AND ((start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?))",
		userID, models.LeaveRejected, req.StartDate, req.StartDate, req.EndDate, req.EndDate).Find(&overlappingLeaves).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check for overlapping leaves")
	}

	if len(overlappingLeaves) > 0 {
		return echo.NewHTTPError(http.StatusConflict, "Leave request overlaps with existing leave")
	}

	leave := models.Leave{
		UserID:    userID,
		Type:      req.Type,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Days:      req.Days,
		Reason:    req.Reason,
		Status:    models.LeavePending,
	}

	if err := h.db.Create(&leave).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create leave request")
	}

	return c.JSON(http.StatusCreated, leave)
}

func (h *LeaveHandler) GetLeaves(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	userRole := c.Get("user_role").(string)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := h.db.Model(&models.Leave{}).Preload("User").Order("created_at DESC")

	if userRole == "admin" || userRole == "manager" {
		status := c.QueryParam("status")
		if status != "" {
			query = query.Where("status = ?", status)
		}
		
		requestUserID := c.QueryParam("user_id")
		if requestUserID != "" {
			query = query.Where("user_id = ?", requestUserID)
		}
	} else {
		query = query.Where("user_id = ?", userID)
	}

	var leaves []models.Leave
	if err := query.Offset(offset).Limit(limit).Find(&leaves).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve leave requests")
	}

	var total int64
	countQuery := h.db.Model(&models.Leave{})
	if userRole == "admin" || userRole == "manager" {
		status := c.QueryParam("status")
		if status != "" {
			countQuery = countQuery.Where("status = ?", status)
		}
		requestUserID := c.QueryParam("user_id")
		if requestUserID != "" {
			countQuery = countQuery.Where("user_id = ?", requestUserID)
		}
	} else {
		countQuery = countQuery.Where("user_id = ?", userID)
	}
	countQuery.Count(&total)

	response := map[string]interface{}{
		"data":     leaves,
		"page":     page,
		"limit":    limit,
		"total":    total,
		"has_next": int64(page*limit) < total,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *LeaveHandler) UpdateLeaveStatus(c echo.Context) error {
	userRole := c.Get("user_role").(string)
	if userRole != "admin" && userRole != "manager" {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
	}

	leaveID, err := strconv.ParseUint(c.Param("leaveId"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid leave ID")
	}

	var req UpdateLeaveStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Status != models.LeaveApproved && req.Status != models.LeaveRejected {
		return echo.NewHTTPError(http.StatusBadRequest, "Status must be approved or rejected")
	}

	var leave models.Leave
	if err := h.db.First(&leave, leaveID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Leave request not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve leave request")
	}

	if leave.Status != models.LeavePending {
		return echo.NewHTTPError(http.StatusBadRequest, "Leave request has already been processed")
	}

	approverID := c.Get("user_id").(uint)
	now := time.Now()

	leave.Status = req.Status
	leave.ApprovedBy = &approverID
	leave.ApprovedAt = &now

	if err := h.db.Save(&leave).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update leave status")
	}

	if err := h.db.Preload("User").Preload("Approver").First(&leave, leaveID).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated leave request")
	}

	return c.JSON(http.StatusOK, leave)
}