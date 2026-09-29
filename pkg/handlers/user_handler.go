package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully.", models.ToUserResponse(user))
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	user, err := h.userService.UpdateProfile(userID, req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile updated successfully.", models.ToUserResponse(user))
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	status := c.Query("status")
	role := c.Query("role")
	sort := c.Query("sort")
	order := c.Query("order")

	users, total, err := h.userService.GetAllUsers(page, perPage, search, status, role, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve users.")
		return
	}

	userResponses := make([]models.UserResponse, 0)
	for _, u := range users {
		userResponses = append(userResponses, models.ToUserResponse(&u))
	}

	lastPage := int(total) / perPage
	if int(total)%perPage > 0 {
		lastPage++
	}

	response.SuccessWithMeta(c, http.StatusOK, "Users retrieved successfully.", userResponses, map[string]interface{}{
		"page":      page,
		"per_page":  perPage,
		"total":     total,
		"last_page": lastPage,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User retrieved successfully.", models.ToUserResponse(user))
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	user, err := h.userService.CreateUser(req, ip, userAgent)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "User created successfully.", models.ToUserResponse(user))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	user, err := h.userService.UpdateUser(id, req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully.", models.ToUserResponse(user))
}

func (h *UserHandler) AdminChangePassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	var req models.AdminChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.userService.AdminChangePassword(id, req, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User password updated successfully.", nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	if !h.canAdminDeleteUser(c, id) {
		response.Forbidden(c, "You do not have permission to delete this user.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.userService.DeleteUser(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User deleted successfully.", nil)
}

func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	var req models.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	user, err := h.userService.AssignRoles(id, req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Roles assigned successfully.", models.ToUserResponse(user))
}

func (h *UserHandler) GetStats(c *gin.Context) {
	count, err := h.userService.GetUserCount()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve stats.")
		return
	}

	response.Success(c, http.StatusOK, "Stats retrieved successfully.", gin.H{
		"total_users": count,
	})
}

func (h *UserHandler) GetTrashedUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	users, total, err := h.userService.GetTrashedUsers(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed users")
		return
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, models.ToUserResponse(&user))
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed users retrieved successfully.", userResponses, meta)
}

func (h *UserHandler) RestoreUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	if !h.canAdminDeleteUser(c, id) {
		response.Forbidden(c, "You do not have permission to restore this user.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.userService.RestoreUser(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User restored successfully.", nil)
}

func (h *UserHandler) ForceDeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	if !h.canAdminDeleteUser(c, id) {
		response.Forbidden(c, "You do not have permission to permanently delete this user.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.userService.ForceDeleteUser(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User permanently deleted.", nil)
}

func (h *UserHandler) canAdminDeleteUser(c *gin.Context, targetUserID uuid.UUID) bool {
	isSuperAdmin, exists := c.Get("is_super_admin")
	if exists && isSuperAdmin.(bool) {
		return true
	}

	targetUser, err := h.userService.GetUserByID(targetUserID)
	if err != nil {
		return true // Let the service handle "not found"
	}

	for _, role := range targetUser.Roles {
		if role.Name == "admin" || role.Name == "super_admin" {
			return false
		}
	}
	return true
}

type BulkRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type BulkStatusRequest struct {
	IDs    []string `json:"ids" binding:"required"`
	Status string   `json:"status" binding:"required"`
}

func (h *UserHandler) BulkDeleteUsers(c *gin.Context) {
	var req BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.canAdminDeleteUser(c, id) {
				if h.userService.DeleteUser(id, ip, userAgent) == nil {
					successCount++
				}
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk delete completed.", gin.H{"success_count": successCount})
}

func (h *UserHandler) BulkRestoreUsers(c *gin.Context) {
	var req BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.canAdminDeleteUser(c, id) {
				if h.userService.RestoreUser(id, ip, userAgent) == nil {
					successCount++
				}
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk restore completed.", gin.H{"success_count": successCount})
}

func (h *UserHandler) BulkForceDeleteUsers(c *gin.Context) {
	var req BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.canAdminDeleteUser(c, id) {
				if h.userService.ForceDeleteUser(id, ip, userAgent) == nil {
					successCount++
				}
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk force delete completed.", gin.H{"success_count": successCount})
}

func (h *UserHandler) BulkUpdateStatus(c *gin.Context) {
	var req BulkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}
	
	status := models.UserStatus(req.Status)
	if status != models.UserStatusActive && status != models.UserStatusPending && status != models.UserStatusInactive && status != models.UserStatusSuspend {
		response.BadRequest(c, "Invalid status value.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			// Get user to verify existence and maybe permissions
			if _, err := h.userService.GetUserByID(id); err == nil {
				updateReq := models.UpdateUserRequest{Status: &status}
				if _, err := h.userService.UpdateUser(id, updateReq, ip, userAgent); err == nil {
					successCount++
				}
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk update status completed.", gin.H{"success_count": successCount})
}
