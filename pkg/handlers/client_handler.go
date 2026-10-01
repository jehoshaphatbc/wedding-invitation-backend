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

type ClientHandler struct {
	clientService *services.ClientService
}

func NewClientHandler(clientService *services.ClientService) *ClientHandler {
	return &ClientHandler{clientService: clientService}
}

func (h *ClientHandler) GetAllClients(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", c.DefaultQuery("per_page", "20"))
	page, _ := strconv.Atoi(pageStr)
	perPage, _ := strconv.Atoi(limitStr)

	search := c.Query("search")
	isTrashed := c.Query("is_trashed") == "true"

	clients, total, err := h.clientService.GetAllClients(page, perPage, search, isTrashed)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve clients: "+err.Error())
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Clients retrieved successfully.", clients, meta)
}

func (h *ClientHandler) GetClient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid client ID.")
		return
	}

	client, err := h.clientService.GetClientByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Client retrieved successfully.", client)
}

func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req models.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	client, err := h.clientService.CreateClient(req, ip, userAgent)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Client created successfully.", client)
}

func (h *ClientHandler) UpdateClient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid client ID.")
		return
	}

	var req models.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	client, err := h.clientService.UpdateClient(id, req, ip, userAgent)
	if err != nil {
		if err.Error() == "client not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Client updated successfully.", client)
}

func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid client ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.clientService.DeleteClient(id, ip, userAgent); err != nil {
		if err.Error() == "client not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete client: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Client deleted successfully (moved to trash).", nil)
}

func (h *ClientHandler) RestoreClient(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		// Could be passed via JSON body { "id": "..." } or query
		var body struct {
			ID uuid.UUID `json:"id"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.ID != uuid.Nil {
			idStr = body.ID.String()
		}
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid client ID for restore.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.clientService.RestoreClient(id, ip, userAgent); err != nil {
		response.InternalServerError(c, "Failed to restore client: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Client restored successfully.", nil)
}

func (h *ClientHandler) ForceDeleteClient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid client ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.clientService.ForceDeleteClient(id, ip, userAgent); err != nil {
		response.InternalServerError(c, "Failed to permanently delete client: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Client permanently deleted.", nil)
}

func (h *ClientHandler) GetTrashedClients(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("per_page", "20")))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	clients, total, err := h.clientService.GetTrashedClients(page, limit, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed clients: "+err.Error())
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": limit,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed clients retrieved successfully.", clients, meta)
}

func (h *ClientHandler) BulkDeleteClients(c *gin.Context) {
	var req models.ClientBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	successCount := 0
	for _, id := range req.IDs {
		if err := h.clientService.DeleteClient(id, ip, userAgent); err == nil {
			successCount++
		}
	}

	response.Success(c, http.StatusOK, "Bulk delete completed.", gin.H{"success_count": successCount})
}

func (h *ClientHandler) BulkRestoreClients(c *gin.Context) {
	var req models.ClientBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	successCount := 0
	for _, id := range req.IDs {
		if err := h.clientService.RestoreClient(id, ip, userAgent); err == nil {
			successCount++
		}
	}

	response.Success(c, http.StatusOK, "Bulk restore completed.", gin.H{"success_count": successCount})
}

func (h *ClientHandler) BulkForceDeleteClients(c *gin.Context) {
	var req models.ClientBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	successCount := 0
	for _, id := range req.IDs {
		if err := h.clientService.ForceDeleteClient(id, ip, userAgent); err == nil {
			successCount++
		}
	}

	response.Success(c, http.StatusOK, "Bulk force delete completed.", gin.H{"success_count": successCount})
}
