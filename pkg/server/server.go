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
	packageRepo := repositories.NewPackageRepository(db)
	templateRepo := repositories.NewTemplateRepository(db)
	featureRepo := repositories.NewFeatureRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	invitationRepo := repositories.NewInvitationRepository(db)
	guestRepo := repositories.NewGuestRepository(db)

	permissionRepo := repositories.NewPermissionRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	companySettingRepo := repositories.NewCompanySettingRepository(db)

	authService := services.NewAuthService(userRepo, refreshTokenRepo, passwordResetRepo, emailVerificationRepo, roleRepo, profileRepo, auditRepo, jwtManager, cfg)
	userService := services.NewUserService(userRepo, roleRepo, profileRepo, auditRepo, refreshTokenRepo)
	roleService := services.NewRoleService(roleRepo, permissionRepo, auditRepo)
	packageService := services.NewPackageService(packageRepo, auditRepo)
	templateService := services.NewTemplateService(templateRepo, auditRepo)
	featureService := services.NewFeatureService(featureRepo, auditRepo)
	clientService := services.NewClientService(clientRepo, auditRepo)
	orderService := services.NewOrderService(orderRepo, clientRepo, packageRepo, invitationRepo, auditRepo)
	clientPortalService := services.NewClientPortalService(orderRepo, invitationRepo, auditRepo)
	clientGuestService := services.NewClientGuestService(guestRepo, invitationRepo, auditRepo)

	companySettingService := services.NewCompanySettingService(companySettingRepo, auditRepo)
	blobService := blob.NewBlobService(cfg)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	roleHandler := handlers.NewRoleHandler(roleService)
	packageHandler := handlers.NewPackageHandler(packageService)
	templateHandler := handlers.NewTemplateHandler(templateService, blobService)
	featureHandler := handlers.NewFeatureHandler(featureService)
	clientHandler := handlers.NewClientHandler(clientService)
	orderHandler := handlers.NewOrderHandler(orderService)
	clientPortalHandler := handlers.NewClientPortalHandler(clientPortalService, orderRepo)
	clientGuestHandler := handlers.NewClientGuestHandler(clientGuestService, orderRepo)

	companySettingHandler := handlers.NewCompanySettingHandler(companySettingService, blobService)
	uploadHandler := handlers.NewUploadHandler(blobService, cfg)
	seederHandler := handlers.NewSeederHandler(db, cfg)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if reqHeaders := c.GetHeader("Access-Control-Request-Headers"); reqHeaders != "" {
			c.Header("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Seeder-Secret, X-Form-Token, X-Client-Token, X-Requested-With, X-CSRF-Token")
		}
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
		// Company Settings (SuperAdmin Only)
		admin.GET("/company-settings", middleware.SuperAdminOnly(), companySettingHandler.AdminGet)
		admin.PUT("/company-settings", middleware.SuperAdminOnly(), companySettingHandler.Update)

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

		// Roles (SuperAdmin Only)
		admin.GET("/roles", middleware.SuperAdminOnly(), roleHandler.GetAllRoles)
		admin.POST("/roles/bulk-delete", middleware.SuperAdminOnly(), roleHandler.BulkDeleteRoles)
		admin.POST("/roles/bulk-restore", middleware.SuperAdminOnly(), roleHandler.BulkRestoreRoles)
		admin.POST("/roles/bulk-force-delete", middleware.SuperAdminOnly(), roleHandler.BulkForceDeleteRoles)
		admin.GET("/roles/trash", middleware.SuperAdminOnly(), roleHandler.GetTrashedRoles)
		admin.POST("/roles", middleware.SuperAdminOnly(), roleHandler.CreateRole)
		admin.GET("/roles/:id", middleware.SuperAdminOnly(), roleHandler.GetRole)
		admin.PATCH("/roles/:id", middleware.SuperAdminOnly(), roleHandler.UpdateRole)
		admin.DELETE("/roles/:id", middleware.SuperAdminOnly(), roleHandler.DeleteRole)
		admin.POST("/roles/:id/restore", middleware.SuperAdminOnly(), roleHandler.RestoreRole)
		admin.DELETE("/roles/:id/force", middleware.SuperAdminOnly(), roleHandler.ForceDeleteRole)
		admin.PUT("/roles/:id/permissions", middleware.SuperAdminOnly(), roleHandler.AssignPermissions)

		// Permissions (SuperAdmin Only)
		admin.GET("/permissions", middleware.SuperAdminOnly(), roleHandler.GetAllPermissions)
		admin.POST("/permissions", middleware.SuperAdminOnly(), roleHandler.CreatePermission)
		admin.PATCH("/permissions/:id", middleware.SuperAdminOnly(), roleHandler.UpdatePermission)
		admin.DELETE("/permissions/:id", middleware.SuperAdminOnly(), roleHandler.DeletePermission)

		admin.GET("/packages", packageHandler.GetAllPackages)
		admin.POST("/packages/bulk-delete", packageHandler.BulkDeletePackages)
		admin.POST("/packages/bulk-restore", middleware.SuperAdminOnly(), packageHandler.BulkRestorePackages)
		admin.POST("/packages/bulk-force-delete", middleware.SuperAdminOnly(), packageHandler.BulkForceDeletePackages)
		admin.GET("/packages/trash", middleware.SuperAdminOnly(), packageHandler.GetTrashedPackages)
		admin.POST("/packages", packageHandler.CreatePackage)
		admin.GET("/packages/:id", packageHandler.GetPackage)
		admin.PATCH("/packages/:id", packageHandler.UpdatePackage)
		admin.DELETE("/packages/:id", packageHandler.DeletePackage)
		admin.POST("/packages/:id/restore", middleware.SuperAdminOnly(), packageHandler.RestorePackage)
		admin.DELETE("/packages/:id/force", middleware.SuperAdminOnly(), packageHandler.ForceDeletePackage)

		admin.GET("/templates", templateHandler.GetAllTemplates)
		admin.POST("/templates/bulk-delete", templateHandler.BulkDeleteTemplates)
		admin.POST("/templates/bulk-restore", middleware.SuperAdminOnly(), templateHandler.BulkRestoreTemplates)
		admin.POST("/templates/bulk-force-delete", middleware.SuperAdminOnly(), templateHandler.BulkForceDeleteTemplates)
		admin.GET("/templates/trash", middleware.SuperAdminOnly(), templateHandler.GetTrashedTemplates)
		admin.POST("/templates", templateHandler.CreateTemplate)
		admin.GET("/templates/:id", templateHandler.GetTemplate)
		admin.PUT("/templates/:id", templateHandler.UpdateTemplate)
		admin.PATCH("/templates/:id", templateHandler.UpdateTemplate)
		admin.DELETE("/templates/:id", templateHandler.DeleteTemplate)
		admin.POST("/templates/:id/restore", middleware.SuperAdminOnly(), templateHandler.RestoreTemplate)
		admin.POST("/templates/restore", middleware.SuperAdminOnly(), templateHandler.RestoreTemplate)
		admin.DELETE("/templates/:id/force", middleware.SuperAdminOnly(), templateHandler.ForceDeleteTemplate)

		// Features (Schema-Driven UI Master Data)
		admin.GET("/features", featureHandler.GetAllFeatures)
		admin.POST("/features", featureHandler.CreateFeature)
		admin.GET("/features/:id", featureHandler.GetFeature)
		admin.PATCH("/features/:id", featureHandler.UpdateFeature)
		admin.DELETE("/features/:id", featureHandler.DeleteFeature)

		// Clients
		admin.GET("/clients", clientHandler.GetAllClients)
		admin.POST("/clients/bulk-delete", clientHandler.BulkDeleteClients)
		admin.POST("/clients/bulk-restore", middleware.SuperAdminOnly(), clientHandler.BulkRestoreClients)
		admin.POST("/clients/bulk-force-delete", middleware.SuperAdminOnly(), clientHandler.BulkForceDeleteClients)
		admin.GET("/clients/trash", middleware.SuperAdminOnly(), clientHandler.GetTrashedClients)
		admin.POST("/clients", clientHandler.CreateClient)
		admin.GET("/clients/:id", clientHandler.GetClient)
		admin.PUT("/clients/:id", clientHandler.UpdateClient)
		admin.PATCH("/clients/:id", clientHandler.UpdateClient)
		admin.DELETE("/clients/:id", clientHandler.DeleteClient)
		admin.POST("/clients/:id/restore", middleware.SuperAdminOnly(), clientHandler.RestoreClient)
		admin.POST("/clients/restore", middleware.SuperAdminOnly(), clientHandler.BulkRestoreClients)
		admin.DELETE("/clients/:id/force", middleware.SuperAdminOnly(), clientHandler.ForceDeleteClient)

		// Orders
		admin.GET("/orders", orderHandler.GetAllOrders)
		admin.POST("/orders/bulk-delete", orderHandler.BulkDeleteOrders)
		admin.POST("/orders/bulk-restore", middleware.SuperAdminOnly(), orderHandler.BulkRestoreOrders)
		admin.POST("/orders/bulk-force-delete", middleware.SuperAdminOnly(), orderHandler.BulkForceDeleteOrders)
		admin.GET("/orders/trash", middleware.SuperAdminOnly(), orderHandler.GetTrashedOrders)
		admin.GET("/orders/:id", orderHandler.GetOrder)
		admin.PUT("/orders/:id", orderHandler.UpdateOrder)
		admin.PATCH("/orders/:id", orderHandler.UpdateOrder)
		admin.DELETE("/orders/:id", orderHandler.DeleteOrder)
		admin.POST("/orders/:id/restore", middleware.SuperAdminOnly(), orderHandler.RestoreOrder)
		admin.POST("/orders/restore", middleware.SuperAdminOnly(), orderHandler.BulkRestoreOrders)
		admin.DELETE("/orders/:id/force", middleware.SuperAdminOnly(), orderHandler.ForceDeleteOrder)
	}

	// Public / Client Accessible Features & Templates List
	api.GET("/features", featureHandler.GetAllFeatures)
	api.GET("/templates", templateHandler.GetAllTemplates)
	api.GET("/templates/:id", templateHandler.GetTemplate)
	r.GET("/api/templates", templateHandler.GetAllTemplates)
	r.GET("/api/templates/:id", templateHandler.GetTemplate)

	// Public Checkout & Webhooks (accessible via both /api/v1 and /api prefix)
	api.POST("/checkout", orderHandler.Checkout)
	api.POST("/webhook/payment", orderHandler.PaymentWebhook)

	r.POST("/api/checkout", orderHandler.Checkout)
	r.POST("/api/webhook/payment", orderHandler.PaymentWebhook)

	// Direct /api/admin alias for clients & orders if called without /v1
	apiAdmin := r.Group("/api/admin")
	apiAdmin.Use(middleware.AuthMiddleware(jwtManager))
	apiAdmin.Use(middleware.OwnershipOrAdminMiddleware(userRepo))
	apiAdmin.Use(middleware.AdminOnly())
	apiAdmin.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		apiAdmin.GET("/clients", clientHandler.GetAllClients)
		apiAdmin.GET("/clients/:id", clientHandler.GetClient)
		apiAdmin.GET("/orders", orderHandler.GetAllOrders)
		apiAdmin.GET("/orders/:id", orderHandler.GetOrder)

		apiAdmin.GET("/templates", templateHandler.GetAllTemplates)
		apiAdmin.GET("/templates/:id", templateHandler.GetTemplate)
		apiAdmin.POST("/templates", templateHandler.CreateTemplate)
		apiAdmin.PUT("/templates/:id", templateHandler.UpdateTemplate)
		apiAdmin.PATCH("/templates/:id", templateHandler.UpdateTemplate)
		apiAdmin.DELETE("/templates/:id", templateHandler.DeleteTemplate)
		apiAdmin.POST("/templates/:id/restore", middleware.SuperAdminOnly(), templateHandler.RestoreTemplate)
		apiAdmin.POST("/templates/restore", middleware.SuperAdminOnly(), templateHandler.RestoreTemplate)
		apiAdmin.POST("/templates/bulk-delete", templateHandler.BulkDeleteTemplates)
		apiAdmin.POST("/templates/bulk-restore", middleware.SuperAdminOnly(), templateHandler.BulkRestoreTemplates)
		apiAdmin.GET("/templates/trash", middleware.SuperAdminOnly(), templateHandler.GetTrashedTemplates)
	}

	// HTTP Seeder Endpoint (protected by SEEDER_SECRET)
	api.GET("/seeder", seederHandler.Execute)
	api.POST("/seeder", seederHandler.Execute)
	r.GET("/api/seeder", seederHandler.Execute)
	r.POST("/api/seeder", seederHandler.Execute)

	// Client Portal (Magic Link - Form Token Auth)
	clientAuthMiddleware := middleware.ClientFormTokenMiddleware(orderRepo)

	clientPortal := api.Group("/client")
	clientPortal.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		clientPortal.GET("/auth-verify", clientPortalHandler.AuthVerify)
		clientPortal.POST("/invitation", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
		clientPortal.PUT("/invitation", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
		clientPortal.GET("/invitation", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)

		// Guests Management
		clientGuests := clientPortal.Group("/guests")
		clientGuests.Use(clientAuthMiddleware)
		{
			clientGuests.GET("", clientGuestHandler.GetGuests)
			clientGuests.POST("/bulk", clientGuestHandler.BulkCreateGuests)
			clientGuests.PUT("/:id", clientGuestHandler.UpdateGuest)
			clientGuests.DELETE("/:id", clientGuestHandler.DeleteGuest)
		}
	}

	// Alias: /api/v1/invitation/setup
	apiInvitationSetup := api.Group("/invitation")
	apiInvitationSetup.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		apiInvitationSetup.POST("/setup", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
		apiInvitationSetup.PUT("/setup", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
	}

	// Alias: /api/client (without /v1)
	clientPortalAlias := r.Group("/api/client")
	clientPortalAlias.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		clientPortalAlias.GET("/auth-verify", clientPortalHandler.AuthVerify)
		clientPortalAlias.POST("/invitation", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
		clientPortalAlias.PUT("/invitation", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
		clientPortalAlias.GET("/invitation", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)

		// Guests Management
		clientGuestsAlias := clientPortalAlias.Group("/guests")
		clientGuestsAlias.Use(clientAuthMiddleware)
		{
			clientGuestsAlias.GET("", clientGuestHandler.GetGuests)
			clientGuestsAlias.POST("/bulk", clientGuestHandler.BulkCreateGuests)
			clientGuestsAlias.PUT("/:id", clientGuestHandler.UpdateGuest)
			clientGuestsAlias.DELETE("/:id", clientGuestHandler.DeleteGuest)
		}
	}

	// Alias: /api/invitation/setup (without /v1)
	rInvitationSetup := r.Group("/api/invitation")
	rInvitationSetup.Use(middleware.RateLimitMiddleware(generalRateLimiter))
	{
		rInvitationSetup.POST("/setup", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
		rInvitationSetup.PUT("/setup", clientAuthMiddleware, clientPortalHandler.UpdateInvitation)
	}

	return r
}

func apiGroup(r *gin.Engine) *gin.RouterGroup {
	return r.Group("/api/v1")
}
