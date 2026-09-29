package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

type TemplateHandler struct {
	templateService *services.TemplateService
}

func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req services.TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: " + err.Error())
		return
	}

	template, err := h.templateService.CreateTemplate(req)
	if err != nil {
		response.InternalServerError(c, "Failed to create template: " + err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Template created successfully.", template)
}

func (h *TemplateHandler) GetAllTemplates(c *gin.Context) {
	templates, err := h.templateService.GetAllTemplates()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve templates: " + err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Templates retrieved successfully.", templates)
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: " + err.Error())
		return
	}

	template, err := h.templateService.UpdateTemplate(id, req)
	if err != nil {
		if err.Error() == "template not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update template: " + err.Error())
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

	if err := h.templateService.DeleteTemplate(id); err != nil {
		if err.Error() == "template not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete template: " + err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Template deleted successfully.", nil)
}
