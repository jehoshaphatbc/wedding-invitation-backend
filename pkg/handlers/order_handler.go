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

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// Public Checkout Endpoint
func (h *OrderHandler) Checkout(c *gin.Context) {
	var req models.PublicCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	order, client, pkg, err := h.orderService.Checkout(req, ip, userAgent)
	if err != nil {
		if err.Error() == "package not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	resData := gin.H{
		"order_id":       order.ID,
		"invoice_number": order.InvoiceNumber,
		"total_amount":   order.TotalAmount,
		"status":         order.Status,
		"payment_url":    order.PaymentURL,
		"client": gin.H{
			"id":       client.ID,
			"name":     client.Name,
			"email":    client.Email,
			"whatsapp": client.Whatsapp,
		},
		"package": gin.H{
			"id":    pkg.ID,
			"name":  pkg.Name,
			"price": pkg.Price,
		},
	}

	response.Success(c, http.StatusCreated, "Checkout initiated successfully.", resData)
}

// Public Payment Webhook Endpoint
func (h *OrderHandler) PaymentWebhook(c *gin.Context) {
	var req models.PaymentWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid webhook payload: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	order, invitation, err := h.orderService.HandlePaymentWebhook(req, ip, userAgent)
	if err != nil {
		if err.Error() == "order not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	resData := gin.H{
		"order_id":       order.ID,
		"invoice_number": order.InvoiceNumber,
		"status":         order.Status,
		"form_token":     order.FormToken,
		"scanner_token":  order.ScannerToken,
	}

	if invitation != nil {
		resData["invitation_id"] = invitation.ID
	}

	response.Success(c, http.StatusOK, "Payment webhook processed successfully.", resData)
}

// Admin Standard Order Endpoints
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", c.DefaultQuery("per_page", "20"))
	page, _ := strconv.Atoi(pageStr)
	perPage, _ := strconv.Atoi(limitStr)

	search := c.Query("search")
	status := c.Query("status")
	isTrashed := c.Query("is_trashed") == "true"

	orders, total, err := h.orderService.GetAllOrders(page, perPage, search, status, isTrashed)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve orders: "+err.Error())
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Orders retrieved successfully.", orders, meta)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID.")
		return
	}

	order, err := h.orderService.GetOrderByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Order retrieved successfully.", order)
}

func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID.")
		return
	}

	var req models.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	order, err := h.orderService.UpdateOrder(id, req, ip, userAgent)
	if err != nil {
		if err.Error() == "order not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Order updated successfully.", order)
}

func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.orderService.DeleteOrder(id, ip, userAgent); err != nil {
		if err.Error() == "order not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete order: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Order deleted successfully (moved to trash).", nil)
}

func (h *OrderHandler) RestoreOrder(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		var body struct {
			ID uuid.UUID `json:"id"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.ID != uuid.Nil {
			idStr = body.ID.String()
		}
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid order ID for restore.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.orderService.RestoreOrder(id, ip, userAgent); err != nil {
		response.InternalServerError(c, "Failed to restore order: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Order restored successfully.", nil)
}

func (h *OrderHandler) ForceDeleteOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.orderService.ForceDeleteOrder(id, ip, userAgent); err != nil {
		response.InternalServerError(c, "Failed to permanently delete order: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Order permanently deleted.", nil)
}

func (h *OrderHandler) GetTrashedOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("per_page", "20")))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	orders, total, err := h.orderService.GetTrashedOrders(page, limit, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed orders: "+err.Error())
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": limit,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed orders retrieved successfully.", orders, meta)
}

func (h *OrderHandler) BulkDeleteOrders(c *gin.Context) {
	var req models.OrderBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	successCount := 0
	for _, id := range req.IDs {
		if err := h.orderService.DeleteOrder(id, ip, userAgent); err == nil {
			successCount++
		}
	}

	response.Success(c, http.StatusOK, "Bulk delete completed.", gin.H{"success_count": successCount})
}

func (h *OrderHandler) BulkRestoreOrders(c *gin.Context) {
	var req models.OrderBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	successCount := 0
	for _, id := range req.IDs {
		if err := h.orderService.RestoreOrder(id, ip, userAgent); err == nil {
			successCount++
		}
	}

	response.Success(c, http.StatusOK, "Bulk restore completed.", gin.H{"success_count": successCount})
}

func (h *OrderHandler) BulkForceDeleteOrders(c *gin.Context) {
	var req models.OrderBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body: "+err.Error())
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	successCount := 0
	for _, id := range req.IDs {
		if err := h.orderService.ForceDeleteOrder(id, ip, userAgent); err == nil {
			successCount++
		}
	}

	response.Success(c, http.StatusOK, "Bulk force delete completed.", gin.H{"success_count": successCount})
}
