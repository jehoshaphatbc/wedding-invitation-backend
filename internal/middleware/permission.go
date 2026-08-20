package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
)

func PermissionMiddleware(userRepo repositories.UserRepository, requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			response.Unauthorized(c, "Unauthenticated.")
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(userID)
		if err != nil {
			response.Unauthorized(c, "Unauthenticated.")
			c.Abort()
			return
		}

		for _, role := range user.Roles {
			if role.Name == "super_admin" {
				c.Next()
				return
			}
		}

		userPermissions := make(map[string]bool)
		for _, role := range user.Roles {
			for _, perm := range role.Permissions {
				userPermissions[perm.Name] = true
			}
		}

		for _, required := range requiredPermissions {
			if !userPermissions[required] {
				response.Forbidden(c, "You do not have permission to perform this action.")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func OwnershipOrAdminMiddleware(userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			response.Unauthorized(c, "Unauthenticated.")
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(userID)
		if err != nil {
			response.Unauthorized(c, "Unauthenticated.")
			c.Abort()
			return
		}

		for _, role := range user.Roles {
			if role.Name == "super_admin" || role.Name == "admin" {
				c.Set("is_admin", true)
				c.Next()
				return
			}
		}

		c.Set("is_admin", false)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			response.Forbidden(c, "You do not have permission to perform this action.")
			c.Abort()
			return
		}

		c.Next()
	}
}
