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

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) GetAllRoles(c *gin.Context) {
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can view roles.")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	roles, total, err := h.roleService.GetAllRoles(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve roles.")
		return
	}

	roleResponses := make([]models.RoleResponse, 0)
	for _, r := range roles {
		roleResponses = append(roleResponses, models.ToRoleResponse(&r))
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Roles retrieved successfully.", roleResponses, meta)
}

func (h *RoleHandler) GetRole(c *gin.Context) {
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can view roles.")
		return
	}

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
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can manage roles.")
		return
	}

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
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can manage roles.")
		return
	}

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

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can delete roles.")
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
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can manage roles.")
		return
	}

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
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can view permissions.")
		return
	}
	permissions, err := h.roleService.GetAllPermissions()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve permissions.")
		return
	}

	response.Success(c, http.StatusOK, "Permissions retrieved successfully.", permissions)
}

func (h *RoleHandler) CreatePermission(c *gin.Context) {
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can manage permissions.")
		return
	}

	var req models.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	perm, err := h.roleService.CreatePermission(req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Permission created successfully.", perm)
}

func (h *RoleHandler) UpdatePermission(c *gin.Context) {
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can manage permissions.")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid permission ID.")
		return
	}

	var req models.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	perm, err := h.roleService.UpdatePermission(id, req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Permission updated successfully.", perm)
}

func (h *RoleHandler) DeletePermission(c *gin.Context) {
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can manage permissions.")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid permission ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.roleService.DeletePermission(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Permission deleted successfully.", nil)
}

func (h *RoleHandler) GetTrashedRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	roles, total, err := h.roleService.GetTrashedRoles(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed roles")
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed roles retrieved successfully.", roles, meta)
}

func (h *RoleHandler) RestoreRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can restore roles.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.roleService.RestoreRole(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Role restored successfully.", nil)
}

func (h *RoleHandler) ForceDeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can permanently delete roles.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.roleService.ForceDeleteRole(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Role permanently deleted.", nil)
}

func (h *RoleHandler) isSuperAdmin(c *gin.Context) bool {
	isSuperAdmin, exists := c.Get("is_super_admin")
	return exists && isSuperAdmin.(bool)
}

type RoleBulkRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func (h *RoleHandler) BulkDeleteRoles(c *gin.Context) {
	var req RoleBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can delete roles.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.roleService.DeleteRole(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk delete completed.", gin.H{"success_count": successCount})
}

func (h *RoleHandler) BulkRestoreRoles(c *gin.Context) {
	var req RoleBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can restore roles.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.roleService.RestoreRole(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk restore completed.", gin.H{"success_count": successCount})
}

func (h *RoleHandler) BulkForceDeleteRoles(c *gin.Context) {
	var req RoleBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can permanently delete roles.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.roleService.ForceDeleteRole(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk force delete completed.", gin.H{"success_count": successCount})
}
