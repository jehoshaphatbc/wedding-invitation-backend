package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

type ClientGuestHandler struct {
	guestService *services.ClientGuestService
	orderRepo    repositories.OrderRepository
}

func NewClientGuestHandler(
	guestService *services.ClientGuestService,
	orderRepo repositories.OrderRepository,
) *ClientGuestHandler {
	return &ClientGuestHandler{
		guestService: guestService,
		orderRepo:    orderRepo,
	}
}

// resolveOrder extracts the authenticated order from context or token headers
func (h *ClientGuestHandler) resolveOrder(c *gin.Context) (*models.Order, error) {
	if val, exists := c.Get("order"); exists {
		if o, ok := val.(*models.Order); ok && o != nil {
			return o, nil
		}
	}

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
		return nil, errors.New("unauthorized")
	}

	order, err := h.orderRepo.FindByFormToken(token)
	if err != nil || order == nil {
		return nil, errors.New("unauthorized")
	}

	return order, nil
}

// GetGuests handles GET /api/client/guests
func (h *ClientGuestHandler) GetGuests(c *gin.Context) {
	order, err := h.resolveOrder(c)
	if err != nil {
		response.Unauthorized(c, "Authorization form token is required.")
		return
	}

	data, err := h.guestService.GetGuests(order)
	if err != nil {
		response.InternalServerError(c, "Gagal mengambil data tamu: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Data tamu berhasil diambil", data)
}

// BulkCreateGuests handles POST /api/client/guests/bulk
func (h *ClientGuestHandler) BulkCreateGuests(c *gin.Context) {
	order, err := h.resolveOrder(c)
	if err != nil {
		response.Unauthorized(c, "Authorization form token is required.")
		return
	}

	var req models.BulkCreateGuestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Payload tidak valid. Format wajib: {\"names\": [\"Nama 1\", \"Nama 2\"]}")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	guests, err := h.guestService.BulkCreateGuests(order, req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Data tamu berhasil ditambahkan", guests)
}

// UpdateGuest handles PUT /api/client/guests/:id
func (h *ClientGuestHandler) UpdateGuest(c *gin.Context) {
	order, err := h.resolveOrder(c)
	if err != nil {
		response.Unauthorized(c, "Authorization form token is required.")
		return
	}

	guestIDStr := c.Param("id")
	guestID, err := uuid.Parse(guestIDStr)
	if err != nil {
		response.BadRequest(c, "ID tamu tidak valid")
		return
	}

	var req models.UpdateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Payload update tidak valid: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	guest, err := h.guestService.UpdateGuest(order, guestID, req, ip, userAgent)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Data tamu berhasil diperbarui", guest)
}

// DeleteGuest handles DELETE /api/client/guests/:id
func (h *ClientGuestHandler) DeleteGuest(c *gin.Context) {
	order, err := h.resolveOrder(c)
	if err != nil {
		response.Unauthorized(c, "Authorization form token is required.")
		return
	}

	guestIDStr := c.Param("id")
	guestID, err := uuid.Parse(guestIDStr)
	if err != nil {
		response.BadRequest(c, "ID tamu tidak valid")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.guestService.DeleteGuest(order, guestID, ip, userAgent); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Data tamu berhasil dihapus", nil)
}
