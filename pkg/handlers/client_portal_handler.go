package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

type ClientPortalHandler struct {
	portalService *services.ClientPortalService
	orderRepo     repositories.OrderRepository
}

func NewClientPortalHandler(
	portalService *services.ClientPortalService,
	orderRepo repositories.OrderRepository,
) *ClientPortalHandler {
	return &ClientPortalHandler{
		portalService: portalService,
		orderRepo:     orderRepo,
	}
}

// AuthVerify validates the magic link form_token from query ?token= or Authorization/Client-Token header.
// GET /api/v1/client/auth-verify
func (h *ClientPortalHandler) AuthVerify(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = strings.TrimSpace(parts[1])
			} else if len(parts) == 1 {
				token = strings.TrimSpace(parts[0])
			}
		}
	}

	if token == "" {
		token = c.GetHeader("X-Client-Token")
	}

	if token == "" {
		token = c.GetHeader("X-Form-Token")
	}

	if token == "" {
		response.Unauthorized(c, "Form token is required.")
		return
	}

	data, err := h.portalService.VerifyToken(token)
	if err != nil {
		response.Unauthorized(c, "Invalid or unauthorized magic link token.")
		return
	}

	response.Success(c, http.StatusOK, "Client authentication verified successfully.", data)
}

// UpdateInvitation saves form inputs submitted by the client (groom, bride, event, theme, story, gallery, gift/gifts)
// POST /api/v1/client/invitation & PUT /api/v1/client/invitation
func (h *ClientPortalHandler) UpdateInvitation(c *gin.Context) {
	var order *models.Order

	// Check if order was already set by ClientFormTokenMiddleware
	if val, exists := c.Get("order"); exists {
		if o, ok := val.(*models.Order); ok {
			order = o
		}
	}

	// Fallback to checking Authorization header or token query if middleware wasn't attached
	if order == nil {
		token := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = strings.TrimSpace(parts[1])
			} else if len(parts) == 1 {
				token = strings.TrimSpace(parts[0])
			}
		}
		if token == "" {
			token = c.GetHeader("X-Client-Token")
		}
		if token == "" {
			token = c.GetHeader("X-Form-Token")
		}
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			response.Unauthorized(c, "Authorization form token is required.")
			return
		}

		var err error
		order, err = h.orderRepo.FindByFormToken(token)
		if err != nil || order == nil {
			response.Unauthorized(c, "Invalid or unauthorized magic link token.")
			return
		}
	}

	var req models.UpdateClientInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request payload: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	invitation, err := h.portalService.UpdateInvitation(order.ID, order.ClientID, order.PackageID, req, ip, userAgent)
	if err != nil {
		if strings.Contains(err.Error(), "permanently locked") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update invitation: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Data undangan berhasil disimpan", invitation)
}
