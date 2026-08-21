package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/dashboard"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/database"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/handlers"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/services"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/auth"
	seeds "github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gin.SetMode(cfg.GinMode)
	db := database.Connect(cfg)

	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
		cfg.JWTIssuer,
	)

	seeds.Seed(db, cfg.SuperAdminName, cfg.SuperAdminEmail, cfg.SuperAdminPassword)

	userRepo := repositories.NewUserRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	passwordResetRepo := repositories.NewPasswordResetTokenRepository(db)
	emailVerificationRepo := repositories.NewEmailVerificationTokenRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)

	authService := services.NewAuthService(userRepo, refreshTokenRepo, passwordResetRepo, emailVerificationRepo, roleRepo, profileRepo, auditRepo, jwtManager, cfg)
	userService := services.NewUserService(userRepo, roleRepo, profileRepo, auditRepo)
	roleService := services.NewRoleService(roleRepo, permissionRepo, auditRepo)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	roleHandler := handlers.NewRoleHandler(roleService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, dashboard.HTML)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Wedding Invitation Backend API"})
	})

	api := r.Group("/api/v1")

	authRateLimiter := middleware.NewRateLimiter(10, 1*time.Minute)
	generalRateLimiter := middleware.NewRateLimiter(100, 1*time.Minute)

	authGroup := api.Group("/auth")
	authGroup.Use(middleware.AuthRateLimitMiddleware(authRateLimiter))
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
		authGroup.POST("/verify-email", authHandler.VerifyEmail)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	protected.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		protected.POST("/auth/logout", authHandler.Logout)
		protected.POST("/auth/logout-all", authHandler.LogoutAll)
		protected.POST("/auth/change-password", authHandler.ChangePassword)

		protected.GET("/me", userHandler.GetProfile)
		protected.PATCH("/me", userHandler.UpdateProfile)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.OwnershipOrAdminMiddleware(userRepo))
	admin.Use(middleware.AdminOnly())
	admin.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		admin.GET("/users", userHandler.GetAllUsers)
		admin.POST("/users", userHandler.CreateUser)
		admin.GET("/users/:id", userHandler.GetUser)
		admin.PATCH("/users/:id", userHandler.UpdateUser)
		admin.DELETE("/users/:id", userHandler.DeleteUser)
		admin.PUT("/users/:id/roles", userHandler.AssignRoles)
		admin.GET("/stats", userHandler.GetStats)

		admin.GET("/roles", roleHandler.GetAllRoles)
		admin.POST("/roles", roleHandler.CreateRole)
		admin.GET("/roles/:id", roleHandler.GetRole)
		admin.PATCH("/roles/:id", roleHandler.UpdateRole)
		admin.DELETE("/roles/:id", roleHandler.DeleteRole)
		admin.PUT("/roles/:id/permissions", roleHandler.AssignPermissions)

		admin.GET("/permissions", roleHandler.GetAllPermissions)
	}

	fmt.Printf("Server starting on port %s\n", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
