package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
)

// ClientFormTokenMiddleware validates the client's magic link form_token.
// It checks the Authorization header ("Bearer <form_token>"), query parameter ("?token=..."),
// or header ("X-Form-Token").
func ClientFormTokenMiddleware(orderRepo repositories.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			c.Abort()
			return
		}

		order, err := orderRepo.FindByFormToken(token)
		if err != nil || order == nil {
			response.Unauthorized(c, "Invalid or unauthorized magic link token.")
			c.Abort()
			return
		}

		c.Set("order", order)
		c.Set("order_id", order.ID)
		c.Set("client_id", order.ClientID)
		c.Set("form_token", token)
		c.Next()
	}
}
