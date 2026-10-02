package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

func TestClientGuests_CRUD(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	// 1. Setup Client 1 & Order 1
	pkg := &models.Package{
		Name:  "Paket Diamond",
		Price: 2000000,
		FeaturesConfig: models.FeaturesConfig{
			"has_qr":   true,
			"has_rsvp": true,
		},
	}
	require.NoError(t, tx.Create(pkg).Error)

	client1 := &models.Client{
		Name:     "Budi & Ani",
		Email:    "budi.ani@example.com",
		Whatsapp: "081122334455",
	}
	require.NoError(t, tx.Create(client1).Error)

	formToken1 := "form_" + uuid.New().String()
	order1 := &models.Order{
		InvoiceNumber: "INV-GUEST-001",
		ClientID:      client1.ID,
		PackageID:     pkg.ID,
		TotalAmount:   2000000,
		Status:        models.OrderStatusPaid,
		FormToken:     &formToken1,
	}
	require.NoError(t, tx.Create(order1).Error)

	// Setup Client 2 & Order 2 (for multi-tenant data isolation test)
	client2 := &models.Client{
		Name:     "Joko & Rini",
		Email:    "joko.rini@example.com",
		Whatsapp: "089988776655",
	}
	require.NoError(t, tx.Create(client2).Error)

	formToken2 := "form_" + uuid.New().String()
	order2 := &models.Order{
		InvoiceNumber: "INV-GUEST-002",
		ClientID:      client2.ID,
		PackageID:     pkg.ID,
		TotalAmount:   2000000,
		Status:        models.OrderStatusPaid,
		FormToken:     &formToken2,
	}
	require.NoError(t, tx.Create(order2).Error)

	// -------------------------------------------------------------
	// TEST 1: GET /api/client/guests - Empty List Initially
	// -------------------------------------------------------------
	t.Run("GET /api/client/guests: 401 when token missing", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/client/guests", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GET /api/client/guests: 200 OK empty list initially", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/client/guests", nil)
		req.Header.Set("Authorization", "Bearer "+formToken1)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    models.ClientGuestListResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, int64(0), resp.Data.TotalGuests)
		assert.Equal(t, int64(0), resp.Data.TotalHadir)
		assert.Equal(t, int64(0), resp.Data.TotalTidakHadir)
		assert.Equal(t, int64(0), resp.Data.TotalPending)
		assert.Empty(t, resp.Data.Guests)
	})

	// -------------------------------------------------------------
	// TEST 2: POST /api/client/guests/bulk - Tambah Tamu Massal
	// -------------------------------------------------------------
	var createdGuestIDs []uuid.UUID

	t.Run("POST /api/client/guests/bulk: 201 OK creates bulk guests", func(t *testing.T) {
		payload := models.BulkCreateGuestsRequest{
			Names: []string{"Yosa", "Arthur", "Nanik"},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/client/guests/bulk", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+formToken1)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool           `json:"success"`
			Message string         `json:"message"`
			Data    []models.Guest `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Len(t, resp.Data, 3)

		for _, g := range resp.Data {
			assert.NotEmpty(t, g.ID)
			assert.NotEmpty(t, g.QRToken)
			assert.Equal(t, "pending", g.RSVPStatus)
			assert.Equal(t, 1, g.Pax)
			createdGuestIDs = append(createdGuestIDs, g.ID)
		}
	})

	// -------------------------------------------------------------
	// TEST 3: GET /api/client/guests - Verify counts and items
	// -------------------------------------------------------------
	t.Run("GET /api/client/guests: returns 3 guests with correct aggregations", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/client/guests", nil)
		req.Header.Set("Authorization", "Bearer "+formToken1)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    models.ClientGuestListResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, int64(3), resp.Data.TotalGuests)
		assert.Equal(t, int64(0), resp.Data.TotalHadir)
		assert.Equal(t, int64(0), resp.Data.TotalTidakHadir)
		assert.Equal(t, int64(3), resp.Data.TotalPending)
		assert.Len(t, resp.Data.Guests, 3)
	})

	// -------------------------------------------------------------
	// TEST 4: PUT /api/client/guests/:id - Edit Tamu
	// -------------------------------------------------------------
	t.Run("PUT /api/client/guests/:id: 200 OK updates guest name", func(t *testing.T) {
		require.NotEmpty(t, createdGuestIDs)
		targetID := createdGuestIDs[0]

		newName := "Yosa Pratama"
		newStatus := "hadir"
		payload := models.UpdateGuestRequest{
			Name:       &newName,
			RSVPStatus: &newStatus,
		}
		body, _ := json.Marshal(payload)

		url := fmt.Sprintf("/api/client/guests/%s", targetID.String())
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+formToken1)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool         `json:"success"`
			Message string       `json:"message"`
			Data    models.Guest `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, "Yosa Pratama", resp.Data.Name)
		assert.Equal(t, "hadir", resp.Data.RSVPStatus)
	})

	t.Run("PUT /api/client/guests/:id: 404 when Client 2 attempts to edit Client 1's guest", func(t *testing.T) {
		require.NotEmpty(t, createdGuestIDs)
		targetID := createdGuestIDs[0]

		newName := "Hacker Name"
		payload := models.UpdateGuestRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(payload)

		url := fmt.Sprintf("/api/client/guests/%s", targetID.String())
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+formToken2) // Client 2's token!

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// -------------------------------------------------------------
	// TEST 5: DELETE /api/client/guests/:id - Hapus Tamu
	// -------------------------------------------------------------
	t.Run("DELETE /api/client/guests/:id: 404 when Client 2 attempts to delete Client 1's guest", func(t *testing.T) {
		require.NotEmpty(t, createdGuestIDs)
		targetID := createdGuestIDs[1] // Arthur

		url := fmt.Sprintf("/api/client/guests/%s", targetID.String())
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+formToken2) // Client 2's token!

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("DELETE /api/client/guests/:id: 200 OK deletes guest for owner", func(t *testing.T) {
		require.NotEmpty(t, createdGuestIDs)
		targetID := createdGuestIDs[1] // Arthur

		url := fmt.Sprintf("/api/client/guests/%s", targetID.String())
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+formToken1)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify count is now 2
		reqList, _ := http.NewRequest(http.MethodGet, "/api/client/guests", nil)
		reqList.Header.Set("Authorization", "Bearer "+formToken1)
		wList := httptest.NewRecorder()
		router.ServeHTTP(wList, reqList)

		var respList struct {
			Success bool                         `json:"success"`
			Data    models.ClientGuestListResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(wList.Body.Bytes(), &respList))
		assert.Equal(t, int64(2), respList.Data.TotalGuests)
		assert.Equal(t, int64(1), respList.Data.TotalHadir)      // Yosa Pratama (hadir)
		assert.Equal(t, int64(0), respList.Data.TotalTidakHadir)
		assert.Equal(t, int64(1), respList.Data.TotalPending)    // Nanik (pending)
	})

	// -------------------------------------------------------------
	// TEST 6: /api/v1/client/guests alias also works
	// -------------------------------------------------------------
	t.Run("GET /api/v1/client/guests: alias works", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/client/guests", nil)
		req.Header.Set("Authorization", "Bearer "+formToken1)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
