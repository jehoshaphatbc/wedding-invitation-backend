package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services/blob"
)

type TemplateHandler struct {
	templateService *services.TemplateService
	blobService     *blob.BlobService
}

func NewTemplateHandler(templateService *services.TemplateService, blobService *blob.BlobService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		blobService:     blobService,
	}
}

func (h *TemplateHandler) handleUpload(c *gin.Context) *string {
	if h.blobService == nil {
		return nil
	}

	var fileHeader *multipart.FileHeader
	var err error
	for _, field := range []string{"thumbnail", "thumbnail_url", "file"} {
		fileHeader, err = c.FormFile(field)
		if err == nil && fileHeader != nil {
			break
		}
	}

	if fileHeader != nil && fileHeader.Size > 0 {
		file, err := fileHeader.Open()
		if err != nil {
			return nil
		}
		defer file.Close()

		ext := filepath.Ext(fileHeader.Filename)
		uuidStr := uuid.New().String()
		filename := fmt.Sprintf("uploads/templates/%s-thumbnail%s", uuidStr, ext)

		res, err := h.blobService.Upload(c.Request.Context(), filename, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
		if err == nil {
			return &res.URL
		}
	}
	return nil
}

func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req services.TemplateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if uploadedURL := h.handleUpload(c); uploadedURL != nil {
		req.ThumbnailURL = uploadedURL
	}

	template, err := h.templateService.CreateTemplate(req)
	if err != nil {
		response.InternalServerError(c, "Failed to create template: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Template created successfully.", template)
}

func (h *TemplateHandler) GetAllTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	templates, total, err := h.templateService.GetAllTemplates(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve templates: "+err.Error())
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Templates retrieved successfully.", templates, meta)
}

func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid template ID.")
		return
	}

	template, err := h.templateService.GetTemplateByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Template retrieved successfully.", template)
}

func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid template ID.")
		return
	}

	var req services.TemplateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if uploadedURL := h.handleUpload(c); uploadedURL != nil {
		req.ThumbnailURL = uploadedURL
	}

	template, err := h.templateService.UpdateTemplate(id, req)
	if err != nil {
		if err.Error() == "template not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update template: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Template updated successfully.", template)
}

func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid template ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.templateService.DeleteTemplate(id, ip, userAgent); err != nil {
		if err.Error() == "template not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete template: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Template deleted successfully.", nil)
}

func (h *TemplateHandler) GetTrashedTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	templates, total, err := h.templateService.GetTrashedTemplates(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed templates")
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed templates retrieved successfully.", templates, meta)
}

func (h *TemplateHandler) RestoreTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid template ID.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can restore templates.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.templateService.RestoreTemplate(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Template restored successfully.", nil)
}

func (h *TemplateHandler) ForceDeleteTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid template ID.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can permanently delete templates.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.templateService.ForceDeleteTemplate(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Template permanently deleted.", nil)
}

func (h *TemplateHandler) isSuperAdmin(c *gin.Context) bool {
	isSuperAdmin, exists := c.Get("is_super_admin")
	return exists && isSuperAdmin.(bool)
}

type TemplateBulkRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func (h *TemplateHandler) BulkDeleteTemplates(c *gin.Context) {
	var req TemplateBulkRequest
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
			if h.templateService.DeleteTemplate(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk delete completed.", gin.H{"success_count": successCount})
}

func (h *TemplateHandler) BulkRestoreTemplates(c *gin.Context) {
	var req TemplateBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can restore templates.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.templateService.RestoreTemplate(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk restore completed.", gin.H{"success_count": successCount})
}

func (h *TemplateHandler) BulkForceDeleteTemplates(c *gin.Context) {
	var req TemplateBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body.")
		return
	}

	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can permanently delete templates.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)
	successCount := 0

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			if h.templateService.ForceDeleteTemplate(id, ip, userAgent) == nil {
				successCount++
			}
		}
	}

	response.Success(c, http.StatusOK, "Bulk force delete completed.", gin.H{"success_count": successCount})
}
