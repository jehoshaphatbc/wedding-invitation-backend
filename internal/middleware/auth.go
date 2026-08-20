package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/auth"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
)

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Unauthenticated.")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Unauthenticated.")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Unauthenticated.")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

func GetUserEmail(c *gin.Context) string {
	email, _ := c.Get("user_email")
	if email == nil {
		return ""
	}
	return email.(string)
}

func GetClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}
	if ip == "" {
		ip = c.ClientIP()
	}
	return ip
}

func GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}
