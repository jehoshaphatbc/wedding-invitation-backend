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

type FeatureHandler struct {
	featureService *services.FeatureService
}

func NewFeatureHandler(featureService *services.FeatureService) *FeatureHandler {
	return &FeatureHandler{featureService: featureService}
}

func (h *FeatureHandler) CreateFeature(c *gin.Context) {
	var req models.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	feature, err := h.featureService.CreateFeature(req, ip, userAgent)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Feature created successfully.", feature)
}

func (h *FeatureHandler) GetAllFeatures(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "20")
	page, _ := strconv.Atoi(pageStr)
	perPage, _ := strconv.Atoi(perPageStr)

	if c.Query("all") == "true" {
		perPage = -1
	}

	search := c.Query("search")
	inputType := c.Query("input_type")
	sort := c.Query("sort")
	order := c.Query("order")

	features, total, err := h.featureService.GetAllFeatures(page, perPage, search, inputType, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve features: "+err.Error())
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Features retrieved successfully.", features, meta)
}

func (h *FeatureHandler) GetFeature(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid feature ID.")
		return
	}

	feature, err := h.featureService.GetFeatureByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Feature retrieved successfully.", feature)
}

func (h *FeatureHandler) UpdateFeature(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid feature ID.")
		return
	}

	var req models.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	feature, err := h.featureService.UpdateFeature(id, req, ip, userAgent)
	if err != nil {
		if err.Error() == "feature not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Feature updated successfully.", feature)
}

func (h *FeatureHandler) DeleteFeature(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid feature ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.featureService.DeleteFeature(id, ip, userAgent); err != nil {
		if err.Error() == "feature not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete feature: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Feature deleted successfully.", nil)
}
