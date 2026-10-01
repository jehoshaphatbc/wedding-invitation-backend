package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

func TestHTTPSeeder_Unauthorized(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Fails with 401 when no secret is provided", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/seeder", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "error", resp["status"])
		assert.Contains(t, resp["message"], "Unauthorized")
	})

	t.Run("Fails with 401 when wrong secret is provided", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/seeder?secret=wrong-secret-xyz", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "error", resp["status"])
	})
}

func TestHTTPSeeder_Success(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Success seed via query param ?secret=, transaction commit, and data verification", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/seeder?secret=test-seeder-secret-123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Status  string         `json:"status"`
			Message string         `json:"message"`
			Data    map[string]any `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "Seeder berhasil dieksekusi", resp.Message)
		assert.Equal(t, float64(3), resp.Data["packages_seeded"])
		assert.Equal(t, float64(3), resp.Data["clients_seeded"])
		assert.Equal(t, float64(5), resp.Data["orders_seeded"])
		assert.Equal(t, float64(3), resp.Data["invitations_seeded"])
		assert.Equal(t, float64(10), resp.Data["guests_seeded"])

		// 1. Verify Packages in DB
		var pkgCount int64
		tx.Model(&models.Package{}).Where("name IN ?", []string{"Paket Silver", "Paket Gold", "Paket Platinum"}).Count(&pkgCount)
		assert.GreaterOrEqual(t, pkgCount, int64(3))

		// 2. Verify Clients in DB
		var clientCount int64
		tx.Model(&models.Client{}).Where("email IN ?", []string{
			"budi.pratama@gmail.com",
			"siti.nurhaliza@gmail.com",
			"dimas.setiawan@gmail.com",
		}).Count(&clientCount)
		assert.Equal(t, int64(3), clientCount)

		// 3. Verify Orders in DB (3 paid, 2 unpaid)
		var paidOrderCount, unpaidOrderCount int64
		tx.Model(&models.Order{}).Where("invoice_number LIKE ?", "INV-SEED-%").Where("status = ?", models.OrderStatusPaid).Count(&paidOrderCount)
		tx.Model(&models.Order{}).Where("invoice_number LIKE ?", "INV-SEED-%").Where("status = ?", models.OrderStatusUnpaid).Count(&unpaidOrderCount)
		assert.Equal(t, int64(3), paidOrderCount)
		assert.Equal(t, int64(2), unpaidOrderCount)

		// Verify paid order has form_token, and scanner_token conditionally based on has_qr
		var samplePaidOrder models.Order
		err = tx.Where("invoice_number = ?", "INV-SEED-001").First(&samplePaidOrder).Error
		require.NoError(t, err)
		assert.NotNil(t, samplePaidOrder.FormToken)
		assert.Nil(t, samplePaidOrder.ScannerToken) // has_qr == false -> scanner_token must be nil/NULL

		var platinumPaidOrder models.Order
		err = tx.Where("invoice_number = ?", "INV-SEED-003").First(&platinumPaidOrder).Error
		require.NoError(t, err)
		assert.NotNil(t, platinumPaidOrder.FormToken)
		assert.NotNil(t, platinumPaidOrder.ScannerToken) // has_qr == true -> scanner_token generated

		// 4. Verify Invitations in DB
		var invCount int64
		tx.Model(&models.Invitation{}).Where("slug LIKE ? OR slug LIKE ? OR slug LIKE ?", "%budi-ani%", "%siti-rizky%", "%dimas-putri%").Count(&invCount)
		assert.GreaterOrEqual(t, invCount, int64(3))

		// 5. Verify Guests in DB
		var guestCount int64
		tx.Model(&models.Guest{}).Where("qr_token LIKE ?", "QR-SEED-%").Count(&guestCount)
		assert.Equal(t, int64(10), guestCount)
	})

	t.Run("Idempotency check: Calling seeder twice executes ON CONFLICT DO NOTHING without error", func(t *testing.T) {
		// First call
		req1, _ := http.NewRequest(http.MethodGet, "/api/seeder?secret=test-seeder-secret-123", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		require.Equal(t, http.StatusOK, w1.Code)

		// Second call immediately
		req2, _ := http.NewRequest(http.MethodGet, "/api/seeder?secret=test-seeder-secret-123", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		var resp2 struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w2.Body.Bytes(), &resp2)
		require.NoError(t, err)
		assert.Equal(t, "success", resp2.Status)
		assert.Equal(t, "Seeder berhasil dieksekusi", resp2.Message)
	})

	t.Run("Success authorization via X-Seeder-Secret header and Bearer token", func(t *testing.T) {
		// Test via X-Seeder-Secret
		reqHeader, _ := http.NewRequest(http.MethodGet, "/api/seeder", nil)
		reqHeader.Header.Set("X-Seeder-Secret", "test-seeder-secret-123")
		wHeader := httptest.NewRecorder()
		router.ServeHTTP(wHeader, reqHeader)
		assert.Equal(t, http.StatusOK, wHeader.Code)

		// Test via Authorization: Bearer
		reqBearer, _ := http.NewRequest(http.MethodGet, "/api/seeder", nil)
		reqBearer.Header.Set("Authorization", "Bearer test-seeder-secret-123")
		wBearer := httptest.NewRecorder()
		router.ServeHTTP(wBearer, reqBearer)
		assert.Equal(t, http.StatusOK, wBearer.Code)
	})
}
