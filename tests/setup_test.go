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

	api := r.Group("/api/v1")
	{
		api.GET("/features", featureHandler.GetAllFeatures)

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
		}
	}

	return r
}
