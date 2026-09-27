package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
)

type CompanySettingHandler struct {
	service *services.CompanySettingService
}

func NewCompanySettingHandler(service *services.CompanySettingService) *CompanySettingHandler {
	return &CompanySettingHandler{service: service}
}

func (h *CompanySettingHandler) Get(c *gin.Context) {
	setting, err := h.service.GetSettings()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve company settings.")
		return
	}
	response.Success(c, http.StatusOK, "Company settings retrieved.", models.ToCompanySettingResponse(setting))
}

func (h *CompanySettingHandler) Update(c *gin.Context) {
	var req models.UpdateCompanySettingRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	var faviconURL, logoLongURL, logoSquareURL *string

	// Helper function to handle file upload
	handleUpload := func(field string) *string {
		file, err := c.FormFile(field)
		if err == nil {
			ext := filepath.Ext(file.Filename)
			filename := fmt.Sprintf("%s_%d%s", field, time.Now().Unix(), ext)
			path := filepath.Join("uploads", "company", filename)
			if err := c.SaveUploadedFile(file, path); err == nil {
				url := "/" + path
				return &url
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
