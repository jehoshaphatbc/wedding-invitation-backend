package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/auth"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/dashboard"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/handlers"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/middleware"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services/blob"
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
		cfg.JWTIssuer,
	)

	userRepo := repositories.NewUserRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	passwordResetRepo := repositories.NewPasswordResetTokenRepository(db)
	emailVerificationRepo := repositories.NewEmailVerificationTokenRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	companySettingRepo := repositories.NewCompanySettingRepository(db)

	authService := services.NewAuthService(userRepo, refreshTokenRepo, passwordResetRepo, emailVerificationRepo, roleRepo, profileRepo, auditRepo, jwtManager, cfg)
	userService := services.NewUserService(userRepo, roleRepo, profileRepo, auditRepo, refreshTokenRepo)
	roleService := services.NewRoleService(roleRepo, permissionRepo, auditRepo)
	companySettingService := services.NewCompanySettingService(companySettingRepo, auditRepo)
	blobService := blob.NewBlobService(cfg)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	roleHandler := handlers.NewRoleHandler(roleService)
	companySettingHandler := handlers.NewCompanySettingHandler(companySettingService, blobService)
	uploadHandler := handlers.NewUploadHandler(blobService, cfg)

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

	api := apiGroup(r)
	r.Static("/uploads", "./uploads")

	api.GET("/company-settings", companySettingHandler.Get)

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
		authGroup.POST("/resend-verification", authHandler.ResendVerification)
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
		protected.POST("/me/email", authHandler.ChangeEmail)

		protected.POST("/upload", uploadHandler.UploadFile)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.OwnershipOrAdminMiddleware(userRepo))
	admin.Use(middleware.AdminOnly())
	admin.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		admin.PUT("/company-settings", companySettingHandler.Update)

		admin.GET("/users", userHandler.GetAllUsers)
		admin.POST("/users/bulk-delete", userHandler.BulkDeleteUsers)
		admin.POST("/users/bulk-restore", middleware.SuperAdminOnly(), userHandler.BulkRestoreUsers)
		admin.POST("/users/bulk-force-delete", middleware.SuperAdminOnly(), userHandler.BulkForceDeleteUsers)
		admin.POST("/users/bulk-status", userHandler.BulkUpdateStatus)

		admin.GET("/users/trash", middleware.SuperAdminOnly(), userHandler.GetTrashedUsers)

		admin.POST("/users", userHandler.CreateUser)
		admin.GET("/users/:id", userHandler.GetUser)
		admin.PATCH("/users/:id", userHandler.UpdateUser)
		admin.DELETE("/users/:id", userHandler.DeleteUser)
		admin.POST("/users/:id/restore", middleware.SuperAdminOnly(), userHandler.RestoreUser)
		admin.DELETE("/users/:id/force", middleware.SuperAdminOnly(), userHandler.ForceDeleteUser)

		admin.PATCH("/users/:id/password", middleware.SuperAdminOnly(), userHandler.AdminChangePassword)
		admin.PUT("/users/:id/roles", userHandler.AssignRoles)
		admin.GET("/stats", userHandler.GetStats)

		admin.GET("/roles", roleHandler.GetAllRoles)
		admin.POST("/roles/bulk-delete", middleware.SuperAdminOnly(), roleHandler.BulkDeleteRoles)
		admin.POST("/roles/bulk-restore", middleware.SuperAdminOnly(), roleHandler.BulkRestoreRoles)
		admin.POST("/roles/bulk-force-delete", middleware.SuperAdminOnly(), roleHandler.BulkForceDeleteRoles)

		admin.GET("/roles/trash", middleware.SuperAdminOnly(), roleHandler.GetTrashedRoles)

		admin.POST("/roles", roleHandler.CreateRole)
		admin.GET("/roles/:id", roleHandler.GetRole)
		admin.PATCH("/roles/:id", roleHandler.UpdateRole)
		admin.DELETE("/roles/:id", roleHandler.DeleteRole)
		admin.POST("/roles/:id/restore", middleware.SuperAdminOnly(), roleHandler.RestoreRole)
		admin.DELETE("/roles/:id/force", middleware.SuperAdminOnly(), roleHandler.ForceDeleteRole)

		admin.PUT("/roles/:id/permissions", roleHandler.AssignPermissions)

		admin.GET("/permissions", roleHandler.GetAllPermissions)
	}

	return r
}

func apiGroup(r *gin.Engine) *gin.RouterGroup {
	return r.Group("/api/v1")
}
