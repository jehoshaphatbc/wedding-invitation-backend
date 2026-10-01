package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

type SeederHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewSeederHandler(db *gorm.DB, cfg *config.Config) *SeederHandler {
	return &SeederHandler{
		db:  db,
		cfg: cfg,
	}
}

func (h *SeederHandler) Execute(c *gin.Context) {
	// ------------------------------------------------------------------
	// 1. SECURITY CHECK: Verify SEEDER_SECRET from Query or Headers
	// ------------------------------------------------------------------
	providedSecret := c.Query("secret")
	if providedSecret == "" {
		providedSecret = c.GetHeader("X-Seeder-Secret")
	}
	if providedSecret == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			providedSecret = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	expectedSecret := h.cfg.SeederSecret
	if expectedSecret == "" || providedSecret != expectedSecret {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized: Invalid or missing seeder secret.",
		})
		return
	}

	// ------------------------------------------------------------------
	// 2. DATABASE TRANSACTION (Automatic Rollback on error / Commit on success)
	// ------------------------------------------------------------------
	var stats map[string]int

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var seedErr error
		stats, seedErr = seeds.SeedDummyData(tx)
		return seedErr
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Seeder execution failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Seeder berhasil dieksekusi",
		"data": gin.H{
			"templates_seeded":   stats["templates_seeded"],
			"packages_seeded":    stats["packages_seeded"],
			"clients_seeded":     stats["clients_seeded"],
			"orders_seeded":      stats["orders_seeded"],
			"invitations_seeded": stats["invitations_seeded"],
			"guests_seeded":      stats["guests_seeded"],
		},
	})
}
