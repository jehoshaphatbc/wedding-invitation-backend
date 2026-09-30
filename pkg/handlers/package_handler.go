package handlers

import (
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

type PackageHandler struct {
	packageService *services.PackageService
}

func NewPackageHandler(packageService *services.PackageService) *PackageHandler {
	return &PackageHandler{packageService: packageService}
}

func (h *PackageHandler) CreatePackage(c *gin.Context) {
	var req services.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	pkg, err := h.packageService.CreatePackage(req)
	if err != nil {
		response.InternalServerError(c, "Failed to create package: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Package created successfully.", pkg)
}

func (h *PackageHandler) GetAllPackages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	packages, total, err := h.packageService.GetAllPackages(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve packages: "+err.Error())
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Packages retrieved successfully.", packages, meta)
}

func (h *PackageHandler) GetPackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	pkg, err := h.packageService.GetPackageByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package retrieved successfully.", pkg)
}

func (h *PackageHandler) UpdatePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	var req services.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	pkg, err := h.packageService.UpdatePackage(id, req)
	if err != nil {
		if err.Error() == "package not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update package: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package updated successfully.", pkg)
}

func (h *PackageHandler) DeletePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.packageService.DeletePackage(id, ip, userAgent); err != nil {
		if err.Error() == "package not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete package: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package deleted successfully.", nil)
}

func (h *PackageHandler) GetTrashedPackages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	packages, total, err := h.packageService.GetTrashedPackages(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed packages")
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed packages retrieved successfully.", packages, meta)
}

func (h *PackageHandler) RestorePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can restore packages.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.packageService.RestorePackage(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package restored successfully.", nil)
}

func (h *PackageHandler) ForceDeletePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can permanently delete packages.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.packageService.ForceDeletePackage(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package permanently deleted.", nil)
}

func (h *PackageHandler) isSuperAdmin(c *gin.Context) bool {
	isSuperAdmin, exists := c.Get("is_super_admin")
	return exists && isSuperAdmin.(bool)
}

type PackageBulkRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func (h *PackageHandler) BulkDeletePackages(c *gin.Context) {
	var req PackageBulkRequest
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
			if h.packageService.DeletePackage(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk delete completed.", gin.H{"success_count": successCount})
}

func (h *PackageHandler) BulkRestorePackages(c *gin.Context) {
	var req PackageBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can restore packages.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.packageService.RestorePackage(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk restore completed.", gin.H{"success_count": successCount})
}

func (h *PackageHandler) BulkForceDeletePackages(c *gin.Context) {
	var req PackageBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can permanently delete packages.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.packageService.ForceDeletePackage(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk force delete completed.", gin.H{"success_count": successCount})
}

