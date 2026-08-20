package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/services"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
)

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.roleService.GetAllRoles()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve roles.")
		return
	}

	roleResponses := make([]models.RoleResponse, 0)
	for _, r := range roles {
		roleResponses = append(roleResponses, models.ToRoleResponse(&r))
	}

	response.Success(c, http.StatusOK, "Roles retrieved successfully.", roleResponses)
}

func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	role, err := h.roleService.GetRoleByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Role retrieved successfully.", models.ToRoleResponse(role))
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	role, err := h.roleService.CreateRole(req, ip, userAgent)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Role created successfully.", models.ToRoleResponse(role))
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	role, err := h.roleService.UpdateRole(id, req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Role updated successfully.", models.ToRoleResponse(role))
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.roleService.DeleteRole(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Role deleted successfully.", nil)
}

func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	var req models.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	role, err := h.roleService.AssignPermissions(id, req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Permissions assigned successfully.", models.ToRoleResponse(role))
}

func (h *RoleHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.roleService.GetAllPermissions()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve permissions.")
		return
	}

	response.Success(c, http.StatusOK, "Permissions retrieved successfully.", permissions)
}
