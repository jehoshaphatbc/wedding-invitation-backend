package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services/blob"
)

type CompanySettingHandler struct {
	service     *services.CompanySettingService
	blobService *blob.BlobService
}

func NewCompanySettingHandler(service *services.CompanySettingService, blobService *blob.BlobService) *CompanySettingHandler {
	return &CompanySettingHandler{service: service, blobService: blobService}
}

func (h *CompanySettingHandler) isSuperAdmin(c *gin.Context) bool {
	isSuperAdmin, exists := c.Get("is_super_admin")
	return exists && isSuperAdmin.(bool)
}

func (h *CompanySettingHandler) Get(c *gin.Context) {
	setting, err := h.service.GetSettings()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve company settings.")
		return
	}
	response.Success(c, http.StatusOK, "Company settings retrieved.", models.ToCompanySettingResponse(setting))
}

func (h *CompanySettingHandler) AdminGet(c *gin.Context) {
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can view company settings.")
		return
	}
	h.Get(c)
}

func (h *CompanySettingHandler) Update(c *gin.Context) {
	if !h.isSuperAdmin(c) {
		response.Forbidden(c, "Only superadmin can update company settings.")
		return
	}

	var req models.UpdateCompanySettingRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	var faviconURL, logoLongURL, logoSquareURL *string

	// Helper function to handle file upload to Vercel Blob
	handleUpload := func(field string) *string {
		fileHeader, err := c.FormFile(field)
		if err == nil && fileHeader != nil {
			// Validation (Max 10MB default)
			// maxUploadSize checking is handled by global setup or can be checked here
			file, err := fileHeader.Open()
			if err != nil {
				return nil
			}
			defer file.Close()

			ext := filepath.Ext(fileHeader.Filename)
			uuidStr := uuid.New().String()
			filename := fmt.Sprintf("uploads/company/%s-%s%s", uuidStr, field, ext)

			res, err := h.blobService.Upload(c.Request.Context(), filename, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
			if err == nil {
				return &res.URL
			}
		}
		return nil
	}

	faviconURL = handleUpload("favicon")
	logoLongURL = handleUpload("logo_long")
	logoSquareURL = handleUpload("logo_square")

	setting, err := h.service.UpdateSettings(req, faviconURL, logoLongURL, logoSquareURL, userID, ip, userAgent)
	if err != nil {
		response.InternalServerError(c, "Failed to update company settings.")
		return
	}

	response.Success(c, http.StatusOK, "Company settings updated successfully.", models.ToCompanySettingResponse(setting))
}
