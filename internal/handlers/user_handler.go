package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/services"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
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

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.userService.DeleteUser(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User status updated successfully.", nil)
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
