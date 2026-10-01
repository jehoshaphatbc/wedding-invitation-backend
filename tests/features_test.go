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

func TestCreateFeature(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Success create new feature", func(t *testing.T) {
		payload := models.CreateFeatureRequest{
			FeatureKey:   "test_has_music",
			FeatureName:  "Background Music Player",
			InputType:    "boolean",
			DefaultValue: "true",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/features", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool           `json:"success"`
			Message string         `json:"message"`
			Data    models.Feature `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, "test_has_music", resp.Data.FeatureKey)
		assert.Equal(t, "Background Music Player", resp.Data.FeatureName)
		assert.Equal(t, "boolean", resp.Data.InputType)
		assert.Equal(t, "true", resp.Data.DefaultValue)
		assert.NotEqual(t, uuid.Nil, resp.Data.ID)
	})

	t.Run("Failed create feature when feature_key is duplicate", func(t *testing.T) {
		// Insert first feature
		payload := models.CreateFeatureRequest{
			FeatureKey:   "test_duplicate_key",
			FeatureName:  "First Feature",
			InputType:    "boolean",
			DefaultValue: "false",
		}
		body, _ := json.Marshal(payload)
		req1, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/features", bytes.NewBuffer(body))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		require.Equal(t, http.StatusCreated, w1.Code)

		// Try to insert same feature_key again
		duplicatePayload := models.CreateFeatureRequest{
			FeatureKey:   "test_duplicate_key",
			FeatureName:  "Second Feature With Duplicate Key",
			InputType:    "number",
			DefaultValue: "10",
		}
		dupBody, _ := json.Marshal(duplicatePayload)
		req2, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/features", bytes.NewBuffer(dupBody))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)

		var resp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w2.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "already exists")
	})
}

func TestGetFeatures(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Ensure seeded default features can be fetched", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/features", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool             `json:"success"`
			Message string           `json:"message"`
			Data    []models.Feature `json:"data"`
			Meta    map[string]any   `json:"meta"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.NotEmpty(t, resp.Data)

		// Verify key default features exist
		keySet := make(map[string]bool)
		for _, f := range resp.Data {
			keySet[f.FeatureKey] = true
		}

		assert.True(t, keySet["has_gallery"], "Expected has_gallery in seeded features")
		assert.True(t, keySet["gallery_limit"], "Expected gallery_limit in seeded features")
		assert.True(t, keySet["has_countdown"], "Expected has_countdown in seeded features")
		assert.True(t, keySet["has_rsvp"], "Expected has_rsvp in seeded features")
	})
}

func TestDeleteFeature(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Ensure feature can be deleted with 200 OK", func(t *testing.T) {
		// 1. Create a temporary feature to delete
		payload := models.CreateFeatureRequest{
			FeatureKey:   "test_temp_feature_for_delete",
			FeatureName:  "Temporary Feature",
			InputType:    "boolean",
			DefaultValue: "false",
		}
		body, _ := json.Marshal(payload)
		reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/features", bytes.NewBuffer(body))
		reqCreate.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		router.ServeHTTP(wCreate, reqCreate)
		require.Equal(t, http.StatusCreated, wCreate.Code)

		var createResp struct {
			Data models.Feature `json:"data"`
		}
		_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)
		featureID := createResp.Data.ID

		// 2. Perform DELETE
		reqDelete, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/features/"+featureID.String(), nil)
		wDelete := httptest.NewRecorder()
		router.ServeHTTP(wDelete, reqDelete)

		assert.Equal(t, http.StatusOK, wDelete.Code)

		var deleteResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(wDelete.Body.Bytes(), &deleteResp)
		require.NoError(t, err)
		assert.True(t, deleteResp.Success)
		assert.Contains(t, deleteResp.Message, "deleted successfully")

		// 3. Verify feature is no longer found via GetByID
		reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/features/"+featureID.String(), nil)
		wGet := httptest.NewRecorder()
		router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusNotFound, wGet.Code)
	})
}

// ==========================================
// UNIT TESTS (Service Layer)
// ==========================================

func TestFeatureService_Unit(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	auditRepo := repositories.NewAuditLogRepository(tx)
	featureRepo := repositories.NewFeatureRepository(tx)
	service := services.NewFeatureService(featureRepo, auditRepo)

	t.Run("CreateFeature and duplicate validation in service", func(t *testing.T) {
		req := models.CreateFeatureRequest{
			FeatureKey:   "service_test_key",
			FeatureName:  "Service Test Feature",
			InputType:    "number",
			DefaultValue: "50",
		}

		feat, err := service.CreateFeature(req, "127.0.0.1", "Go-Test")
		require.NoError(t, err)
		assert.NotNil(t, feat)
		assert.Equal(t, "service_test_key", feat.FeatureKey)

		// Second create with same key should return error
		_, errDup := service.CreateFeature(req, "127.0.0.1", "Go-Test")
		assert.Error(t, errDup)
		assert.Contains(t, errDup.Error(), "already exists")
	})

	t.Run("GetFeatureByID in service", func(t *testing.T) {
		req := models.CreateFeatureRequest{
			FeatureKey:   "service_find_key",
			FeatureName:  "Service Find Feature",
			InputType:    "boolean",
			DefaultValue: "true",
		}
		feat, err := service.CreateFeature(req, "127.0.0.1", "Go-Test")
		require.NoError(t, err)

		found, err := service.GetFeatureByID(feat.ID)
		require.NoError(t, err)
		assert.Equal(t, feat.ID, found.ID)
		assert.Equal(t, "service_find_key", found.FeatureKey)
	})
}
