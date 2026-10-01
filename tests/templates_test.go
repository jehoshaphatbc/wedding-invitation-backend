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
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

func TestTemplatesCRUD(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Create Template with validation", func(t *testing.T) {
		// Validation failure: missing name
		invalidPayload := map[string]any{
			"nuxt_component": "TemplateTest",
		}
		body, _ := json.Marshal(invalidPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/admin/templates", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Success creation
		thumb := "https://example.com/thumb.jpg"
		active := true
		validPayload := services.TemplateRequest{
			Name:          "Minimalist Elegance",
			NuxtComponent: "TemplateA",
			ThumbnailURL:  &thumb,
			IsActive:      &active,
		}
		body, _ = json.Marshal(validPayload)
		req, _ = http.NewRequest(http.MethodPost, "/api/admin/templates", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool            `json:"success"`
			Data    models.Template `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Minimalist Elegance", resp.Data.Name)
		assert.Equal(t, "TemplateA", resp.Data.NuxtComponent)
		assert.True(t, resp.Data.IsActive)
		assert.NotNil(t, resp.Data.ThumbnailURL)
		assert.Equal(t, "https://example.com/thumb.jpg", *resp.Data.ThumbnailURL)

		createdID := resp.Data.ID

		// Get By ID
		reqGet, _ := http.NewRequest(http.MethodGet, "/api/admin/templates/"+createdID.String(), nil)
		wGet := httptest.NewRecorder()
		router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusOK, wGet.Code)

		var getResp struct {
			Success bool            `json:"success"`
			Data    models.Template `json:"data"`
		}
		_ = json.Unmarshal(wGet.Body.Bytes(), &getResp)
		assert.Equal(t, createdID, getResp.Data.ID)
		assert.Equal(t, "Minimalist Elegance", getResp.Data.Name)

		// Update (PUT)
		newActive := false
		updatePayload := services.TemplateRequest{
			Name:          "Minimalist Elegance Updated",
			NuxtComponent: "TemplateA2",
			IsActive:      &newActive,
		}
		bodyUpdate, _ := json.Marshal(updatePayload)
		reqUpdate, _ := http.NewRequest(http.MethodPut, "/api/admin/templates/"+createdID.String(), bytes.NewBuffer(bodyUpdate))
		reqUpdate.Header.Set("Content-Type", "application/json")
		wUpdate := httptest.NewRecorder()
		router.ServeHTTP(wUpdate, reqUpdate)
		assert.Equal(t, http.StatusOK, wUpdate.Code)

		var updateResp struct {
			Data models.Template `json:"data"`
		}
		_ = json.Unmarshal(wUpdate.Body.Bytes(), &updateResp)
		assert.Equal(t, "Minimalist Elegance Updated", updateResp.Data.Name)
		assert.False(t, updateResp.Data.IsActive)

		// List filter by is_active=false
		reqFilter, _ := http.NewRequest(http.MethodGet, "/api/admin/templates?is_active=false&search=Updated", nil)
		wFilter := httptest.NewRecorder()
		router.ServeHTTP(wFilter, reqFilter)
		assert.Equal(t, http.StatusOK, wFilter.Code)

		var listResp struct {
			Data []models.Template `json:"data"`
			Meta map[string]any    `json:"meta"`
		}
		_ = json.Unmarshal(wFilter.Body.Bytes(), &listResp)
		assert.NotEmpty(t, listResp.Data)
		assert.Equal(t, "Minimalist Elegance Updated", listResp.Data[0].Name)

		// Soft Delete
		reqDel, _ := http.NewRequest(http.MethodDelete, "/api/admin/templates/"+createdID.String(), nil)
		wDel := httptest.NewRecorder()
		router.ServeHTTP(wDel, reqDel)
		assert.Equal(t, http.StatusOK, wDel.Code)

		// Verify is_trashed=true sees it
		reqTrash, _ := http.NewRequest(http.MethodGet, "/api/admin/templates?is_trashed=true&search=Updated", nil)
		wTrash := httptest.NewRecorder()
		router.ServeHTTP(wTrash, reqTrash)
		assert.Equal(t, http.StatusOK, wTrash.Code)

		var trashResp struct {
			Data []models.Template `json:"data"`
		}
		_ = json.Unmarshal(wTrash.Body.Bytes(), &trashResp)
		assert.NotEmpty(t, trashResp.Data)
		assert.Equal(t, createdID, trashResp.Data[0].ID)

		// Restore
		reqRestore, _ := http.NewRequest(http.MethodPost, "/api/admin/templates/"+createdID.String()+"/restore", nil)
		wRestore := httptest.NewRecorder()
		router.ServeHTTP(wRestore, reqRestore)
		assert.Equal(t, http.StatusOK, wRestore.Code)

		// Verify restored item is active in normal list
		reqNormal, _ := http.NewRequest(http.MethodGet, "/api/admin/templates?search=Updated", nil)
		wNormal := httptest.NewRecorder()
		router.ServeHTTP(wNormal, reqNormal)
		assert.Equal(t, http.StatusOK, wNormal.Code)

		var normalResp struct {
			Data []models.Template `json:"data"`
		}
		_ = json.Unmarshal(wNormal.Body.Bytes(), &normalResp)
		assert.NotEmpty(t, normalResp.Data)
		assert.Equal(t, createdID, normalResp.Data[0].ID)
	})

	t.Run("Bulk Delete and Bulk Restore for Templates", func(t *testing.T) {
		t1 := models.Template{Name: "Bulk T1", NuxtComponent: "Comp1", IsActive: true}
		t2 := models.Template{Name: "Bulk T2", NuxtComponent: "Comp2", IsActive: true}
		require.NoError(t, tx.Create(&t1).Error)
		require.NoError(t, tx.Create(&t2).Error)

		// Bulk delete
		bulkBody, _ := json.Marshal(map[string]any{
			"ids": []uuid.UUID{t1.ID, t2.ID},
		})
		reqBulkDel, _ := http.NewRequest(http.MethodPost, "/api/admin/templates/bulk-delete", bytes.NewBuffer(bulkBody))
		reqBulkDel.Header.Set("Content-Type", "application/json")
		wBulkDel := httptest.NewRecorder()
		router.ServeHTTP(wBulkDel, reqBulkDel)
		assert.Equal(t, http.StatusOK, wBulkDel.Code)

		// Verify in trash
		var countTrashed int64
		tx.Unscoped().Model(&models.Template{}).Where("id IN ? AND deleted_at IS NOT NULL", []uuid.UUID{t1.ID, t2.ID}).Count(&countTrashed)
		assert.Equal(t, int64(2), countTrashed)

		// Bulk restore
		reqBulkRes, _ := http.NewRequest(http.MethodPost, "/api/admin/templates/bulk-restore", bytes.NewBuffer(bulkBody))
		reqBulkRes.Header.Set("Content-Type", "application/json")
		wBulkRes := httptest.NewRecorder()
		router.ServeHTTP(wBulkRes, reqBulkRes)
		assert.Equal(t, http.StatusOK, wBulkRes.Code)

		// Verify restored
		var countActive int64
		tx.Model(&models.Template{}).Where("id IN ?", []uuid.UUID{t1.ID, t2.ID}).Count(&countActive)
		assert.Equal(t, int64(2), countActive)
	})
}
