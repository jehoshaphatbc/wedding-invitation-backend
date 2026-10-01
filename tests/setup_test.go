package tests

import (
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/database"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/handlers"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

var globalDB *gorm.DB

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Load .env from project root
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config for tests: %v", err)
	} else {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Printf("Failed to connect to database for tests: %v", err)
		} else {
			globalDB = db
		}
	}

	code := m.Run()
	os.Exit(code)
}

// setupTx returns a transactional DB instance that rolls back when cleanup is called,
// ensuring the main database is completely untouched after each test.
func setupTx(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	if globalDB == nil {
		t.Skip("Skipping test: Database connection not available")
	}

	tx := globalDB.Begin()
	cleanup := func() {
		tx.Rollback()
	}
	return tx, cleanup
}

// setupTestRouter builds a Gin engine wired with test transaction repositories and handlers.
func setupTestRouter(tx *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	auditRepo := repositories.NewAuditLogRepository(tx)

	featureRepo := repositories.NewFeatureRepository(tx)
	featureService := services.NewFeatureService(featureRepo, auditRepo)
	featureHandler := handlers.NewFeatureHandler(featureService)

	packageRepo := repositories.NewPackageRepository(tx)
	packageService := services.NewPackageService(packageRepo, auditRepo)
	packageHandler := handlers.NewPackageHandler(packageService)

	clientRepo := repositories.NewClientRepository(tx)
	clientService := services.NewClientService(clientRepo, auditRepo)
	clientHandler := handlers.NewClientHandler(clientService)

	orderRepo := repositories.NewOrderRepository(tx)
	invitationRepo := repositories.NewInvitationRepository(tx)
	orderService := services.NewOrderService(orderRepo, clientRepo, packageRepo, invitationRepo, auditRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	api := r.Group("/api/v1")
	{
		api.GET("/features", featureHandler.GetAllFeatures)

		// Public Checkout & Webhooks
		api.POST("/checkout", orderHandler.Checkout)
		api.POST("/webhook/payment", orderHandler.PaymentWebhook)

		admin := api.Group("/admin")
		// Simulate authenticated admin context
		admin.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New())
			c.Set("role", "super_admin")
			c.Next()
		})
		{
			// Features
			admin.GET("/features", featureHandler.GetAllFeatures)
			admin.POST("/features", featureHandler.CreateFeature)
			admin.GET("/features/:id", featureHandler.GetFeature)
			admin.PATCH("/features/:id", featureHandler.UpdateFeature)
			admin.DELETE("/features/:id", featureHandler.DeleteFeature)

			// Packages
			admin.GET("/packages", packageHandler.GetAllPackages)
			admin.POST("/packages", packageHandler.CreatePackage)
			admin.GET("/packages/:id", packageHandler.GetPackage)
			admin.PATCH("/packages/:id", packageHandler.UpdatePackage)
			admin.DELETE("/packages/:id", packageHandler.DeletePackage)

			// Clients
			admin.GET("/clients", clientHandler.GetAllClients)
			admin.POST("/clients/bulk-delete", clientHandler.BulkDeleteClients)
			admin.POST("/clients/bulk-restore", clientHandler.BulkRestoreClients)
			admin.POST("/clients/bulk-force-delete", clientHandler.BulkForceDeleteClients)
			admin.GET("/clients/trash", clientHandler.GetTrashedClients)
			admin.POST("/clients", clientHandler.CreateClient)
			admin.GET("/clients/:id", clientHandler.GetClient)
			admin.PUT("/clients/:id", clientHandler.UpdateClient)
			admin.PATCH("/clients/:id", clientHandler.UpdateClient)
			admin.DELETE("/clients/:id", clientHandler.DeleteClient)
			admin.POST("/clients/:id/restore", clientHandler.RestoreClient)
			admin.POST("/clients/restore", clientHandler.BulkRestoreClients)
			admin.DELETE("/clients/:id/force", clientHandler.ForceDeleteClient)

			// Orders
			admin.GET("/orders", orderHandler.GetAllOrders)
			admin.POST("/orders/bulk-delete", orderHandler.BulkDeleteOrders)
			admin.POST("/orders/bulk-restore", orderHandler.BulkRestoreOrders)
			admin.POST("/orders/bulk-force-delete", orderHandler.BulkForceDeleteOrders)
			admin.GET("/orders/trash", orderHandler.GetTrashedOrders)
			admin.GET("/orders/:id", orderHandler.GetOrder)
			admin.PUT("/orders/:id", orderHandler.UpdateOrder)
			admin.PATCH("/orders/:id", orderHandler.UpdateOrder)
			admin.DELETE("/orders/:id", orderHandler.DeleteOrder)
			admin.POST("/orders/:id/restore", orderHandler.RestoreOrder)
			admin.POST("/orders/restore", orderHandler.BulkRestoreOrders)
			admin.DELETE("/orders/:id/force", orderHandler.ForceDeleteOrder)
		}
	}

	// Public root-level aliases
	r.POST("/api/checkout", orderHandler.Checkout)
	r.POST("/api/webhook/payment", orderHandler.PaymentWebhook)

	// Direct /api/admin aliases
	apiAdmin := r.Group("/api/admin")
	{
		apiAdmin.GET("/clients", clientHandler.GetAllClients)
		apiAdmin.GET("/clients/:id", clientHandler.GetClient)
		apiAdmin.GET("/orders", orderHandler.GetAllOrders)
		apiAdmin.GET("/orders/:id", orderHandler.GetOrder)
	}

	// Seeder handler with test secret
	cfg := &config.Config{SeederSecret: "test-seeder-secret-123"}
	seederHandler := handlers.NewSeederHandler(tx, cfg)
	api.GET("/seeder", seederHandler.Execute)
	api.POST("/seeder", seederHandler.Execute)
	r.GET("/api/seeder", seederHandler.Execute)
	r.POST("/api/seeder", seederHandler.Execute)

	return r
}
