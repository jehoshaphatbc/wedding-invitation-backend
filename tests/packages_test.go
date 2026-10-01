package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

// ==========================================
// INTEGRATION TESTS (API Endpoints & DB)
// ==========================================

func TestCreatePackage(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Success insert package with valid JSONB features_config into PostgreSQL", func(t *testing.T) {
		payload := map[string]any{
			"name":  "Paket Royal Diamond",
			"price": 750000,
			"features_config": map[string]any{
				"has_countdown": true,
				"has_maps":      true,
				"has_rsvp":      true,
				"has_gift":      true,
				"has_gallery":   true,
				"gallery_limit": 20,
				"has_video":     true,
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/packages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool           `json:"success"`
			Message string         `json:"message"`
			Data    models.Package `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, "Paket Royal Diamond", resp.Data.Name)
		assert.Equal(t, float64(750000), resp.Data.Price)
		assert.NotEqual(t, uuid.Nil, resp.Data.ID)

		// Verify JSONB features_config structure in response
		cfg := resp.Data.FeaturesConfig
		assert.Equal(t, true, cfg["has_gallery"])
		assert.Equal(t, float64(20), cfg["gallery_limit"])
		assert.Equal(t, true, cfg["has_video"])

		// Direct PostgreSQL query verification: verify it actually landed in postgres table
		var dbPkg models.Package
		err = tx.First(&dbPkg, "id = ?", resp.Data.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Paket Royal Diamond", dbPkg.Name)
		assert.Equal(t, true, dbPkg.FeaturesConfig["has_gallery"])
	})
}

func TestGetPackageByID(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Ensure JSONB is decoded properly into struct when fetched", func(t *testing.T) {
		// 1. Create a package first
		initialConfig := models.FeaturesConfig{
			"has_countdown": true,
			"has_gallery":   true,
			"gallery_limit": float64(15),
			"has_story":     false,
		}
		pkg := &models.Package{
			Name:           "Paket Silver Elegant",
			Price:          250000,
			FeaturesConfig: initialConfig,
		}
		require.NoError(t, tx.Create(pkg).Error)

		// 2. Fetch via API
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/packages/"+pkg.ID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool           `json:"success"`
			Message string         `json:"message"`
			Data    models.Package `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, pkg.ID, resp.Data.ID)
		assert.Equal(t, "Paket Silver Elegant", resp.Data.Name)

		// 3. Assert JSONB fields decoded properly
		decodedConfig := resp.Data.FeaturesConfig
		require.NotNil(t, decodedConfig)
		assert.Equal(t, true, decodedConfig["has_countdown"])
		assert.Equal(t, true, decodedConfig["has_gallery"])
		assert.Equal(t, float64(15), decodedConfig["gallery_limit"])
		assert.Equal(t, false, decodedConfig["has_story"])
	})
}

func TestUpdatePackage_RemoveFeature(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Simulate FE sending updated features_config with removed feature keys", func(t *testing.T) {
		// 1. Create a package initially containing 'has_gallery' and 'gallery_limit'
		initialConfig := models.FeaturesConfig{
			"has_countdown": true,
			"has_maps":      true,
			"has_gallery":   true,
			"gallery_limit": float64(10),
			"has_qr":        true,
		}
		pkg := &models.Package{
			Name:           "Paket Gold Modifikasi",
			Price:          450000,
			FeaturesConfig: initialConfig,
		}
		require.NoError(t, tx.Create(pkg).Error)

		// Verify initial state has gallery keys
		assert.Contains(t, pkg.FeaturesConfig, "has_gallery")
		assert.Contains(t, pkg.FeaturesConfig, "gallery_limit")

		// 2. Simulate master feature removal:
		// FE now submits features_config WITHOUT 'has_gallery' and 'gallery_limit'
		updatedPayload := services.PackageRequest{
			Name:  "Paket Gold Modifikasi (Updated)",
			Price: 400000,
			FeaturesConfig: models.FeaturesConfig{
				"has_countdown": true,
				"has_maps":      true,
				"has_qr":        true,
			},
		}

		body, _ := json.Marshal(updatedPayload)
		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/admin/packages/"+pkg.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool           `json:"success"`
			Data    models.Package `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		// 3. Verify response no longer contains removed keys
		assert.NotContains(t, resp.Data.FeaturesConfig, "has_gallery")
		assert.NotContains(t, resp.Data.FeaturesConfig, "gallery_limit")
		assert.Equal(t, true, resp.Data.FeaturesConfig["has_countdown"])
		assert.Equal(t, true, resp.Data.FeaturesConfig["has_maps"])

		// 4. Verify PostgreSQL directly: new JSONB completely replaced the old JSONB
		var dbPkg models.Package
		err = tx.First(&dbPkg, "id = ?", pkg.ID).Error
		require.NoError(t, err)

		assert.Equal(t, "Paket Gold Modifikasi (Updated)", dbPkg.Name)
		assert.Equal(t, float64(400000), dbPkg.Price)
		assert.NotContains(t, dbPkg.FeaturesConfig, "has_gallery")
		assert.NotContains(t, dbPkg.FeaturesConfig, "gallery_limit")
		assert.Contains(t, dbPkg.FeaturesConfig, "has_countdown")
		assert.Contains(t, dbPkg.FeaturesConfig, "has_maps")
	})
}

func TestDeletePackage(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Ensure package delete uses soft delete", func(t *testing.T) {
		// 1. Create a package
		pkg := &models.Package{
			Name:  "Paket Yang Akan Dihapus",
			Price: 150000,
			FeaturesConfig: models.FeaturesConfig{
				"has_countdown": true,
			},
		}
		require.NoError(t, tx.Create(pkg).Error)

		// 2. Send DELETE request
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/packages/"+pkg.ID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		// 3. Verify normal query returns ErrRecordNotFound (soft deleted)
		var normalQueryPkg models.Package
		err = tx.First(&normalQueryPkg, "id = ?", pkg.ID).Error
		assert.Error(t, err, "Normal query should not find soft-deleted package")

		// 4. Verify Unscoped query still finds the record with deleted_at IS NOT NULL
		var unscopedPkg models.Package
		err = tx.Unscoped().First(&unscopedPkg, "id = ?", pkg.ID).Error
		require.NoError(t, err, "Unscoped query should find soft-deleted package")
		assert.True(t, unscopedPkg.DeletedAt.Valid, "deleted_at must be valid (Soft Deleted)")
		assert.False(t, unscopedPkg.DeletedAt.Time.IsZero())
	})
}

// ==========================================
// UNIT TESTS (Service Layer)
// ==========================================

func TestPackageService_Unit(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	auditRepo := repositories.NewAuditLogRepository(tx)
	packageRepo := repositories.NewPackageRepository(tx)
	service := services.NewPackageService(packageRepo, auditRepo)

	t.Run("PackageService CreatePackage and GetPackageByID", func(t *testing.T) {
		req := services.PackageRequest{
			Name:  "Unit Test Package",
			Price: 300000,
			FeaturesConfig: models.FeaturesConfig{
				"has_rsvp": true,
				"has_gift": true,
			},
		}

		created, err := service.CreatePackage(req)
		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, "Unit Test Package", created.Name)

		found, err := service.GetPackageByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, true, found.FeaturesConfig["has_rsvp"])
	})

	t.Run("PackageService UpdatePackage overwriting features_config", func(t *testing.T) {
		initial, err := service.CreatePackage(services.PackageRequest{
			Name:  "Initial Unit Package",
			Price: 200000,
			FeaturesConfig: models.FeaturesConfig{
				"has_gallery":   true,
				"gallery_limit": float64(5),
			},
		})
		require.NoError(t, err)

		updated, err := service.UpdatePackage(initial.ID, services.PackageRequest{
			Name:  "Updated Unit Package",
			Price: 250000,
			FeaturesConfig: models.FeaturesConfig{
				"has_video": true,
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Unit Package", updated.Name)
		assert.NotContains(t, updated.FeaturesConfig, "has_gallery")
		assert.Equal(t, true, updated.FeaturesConfig["has_video"])
	})
}
