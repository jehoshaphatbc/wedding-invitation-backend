package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services/blob"
)

type UploadHandler struct {
	blobService *blob.BlobService
	cfg         *config.Config
}

func NewUploadHandler(blobService *blob.BlobService, cfg *config.Config) *UploadHandler {
	return &UploadHandler{
		blobService: blobService,
		cfg:         cfg,
	}
}

func (h *UploadHandler) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.ValidationError(c, "File is required")
		return
	}

	if fileHeader.Size == 0 {
		response.ValidationError(c, "File cannot be empty")
		return
	}

	if fileHeader.Size > h.cfg.MaxUploadSize {
		response.ValidationError(c, fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", h.cfg.MaxUploadSize))
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		response.ValidationError(c, "File extension is required")
		return
	}

	// Basic validation for common web files
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".svg": true, ".gif": true,
		".pdf": true, ".csv": true, ".doc": true, ".docx": true,
	}
	if !allowedExts[ext] {
		response.ValidationError(c, "File type not allowed")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalServerError(c, "Failed to open file")
		return
	}
	defer file.Close()

	uuidStr := uuid.New().String()
	// Default generic entity if not provided, you could optionally read from a form field like 'entity'
	entity := c.PostForm("entity")
	if entity == "" {
		entity = "general"
	}

	filename := fmt.Sprintf("uploads/%s/%s-%s", entity, uuidStr, filepath.Base(fileHeader.Filename))
	contentType := fileHeader.Header.Get("Content-Type")

	res, err := h.blobService.Upload(c.Request.Context(), filename, file, contentType)
	if err != nil {
		response.InternalServerError(c, "Failed to upload file to storage")
		return
	}

	response.Success(c, http.StatusOK, "File uploaded successfully", gin.H{
		"url":         res.URL,
		"pathname":    res.Pathname,
		"contentType": res.ContentType,
	})
}
