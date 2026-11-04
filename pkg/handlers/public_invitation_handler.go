package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

type PublicInvitationHandler struct {
	publicService *services.PublicInvitationService
}

func NewPublicInvitationHandler(publicService *services.PublicInvitationService) *PublicInvitationHandler {
	return &PublicInvitationHandler{
		publicService: publicService,
	}
}

// GetBySlug handles GET /api/v1/invitations/slug/:slug and GET /api/v1/public/invitations/:slug
func (h *PublicInvitationHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		slug = c.Param("id")
	}

	data, err := h.publicService.GetInvitationBySlug(slug)
	if err != nil {
		if err.Error() == "invitation not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"success": false,
				"message": "Invitation not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"success": true,
		"data":    data,
	})
}

// SubmitRSVP handles POST /api/v1/invitations/:id/rsvp
func (h *PublicInvitationHandler) SubmitRSVP(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		id = c.Param("slug")
	}

	var req models.PublicRSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request payload: "+err.Error())
		return
	}

	result, err := h.publicService.SubmitRSVP(id, req)
	if err != nil {
		if err.Error() == "invitation not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"success": false,
				"message": "Invitation not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"success": true,
		"message": "RSVP berhasil dikirim",
		"data":    result,
	})
}

// GetWishes handles GET /api/v1/invitations/:id/wishes
func (h *PublicInvitationHandler) GetWishes(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		id = c.Param("slug")
	}

	wishes, err := h.publicService.GetWishes(id)
	if err != nil {
		if err.Error() == "invitation not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"success": false,
				"message": "Invitation not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if wishes == nil {
		wishes = []models.Wish{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"success": true,
		"total":   len(wishes),
		"data":    wishes,
	})
}
