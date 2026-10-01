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
)

// ====================================================
// 1. PUBLIC CHECKOUT & PAYMENT WEBHOOK TESTS
// ====================================================

func TestPublicCheckout(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	// Create test package
	pkg := &models.Package{
		Name:  "Paket Premium Checkout",
		Price: 500000,
		FeaturesConfig: models.FeaturesConfig{
			"has_countdown": true,
			"has_gallery":   true,
		},
	}
	require.NoError(t, tx.Create(pkg).Error)

	t.Run("Success checkout creates/finds client, creates unpaid order, and returns mockup payment_url", func(t *testing.T) {
		payload := models.PublicCheckoutRequest{
			Name:      "Budi Santoso",
			Email:     "budi.santoso@example.com",
			Whatsapp:  "081234567890",
			PackageID: pkg.ID,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/checkout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				OrderID       uuid.UUID `json:"order_id"`
				InvoiceNumber string    `json:"invoice_number"`
				TotalAmount   float64   `json:"total_amount"`
				Status        string    `json:"status"`
				PaymentURL    string    `json:"payment_url"`
				Client        struct {
					ID       uuid.UUID `json:"id"`
					Name     string    `json:"name"`
					Email    string    `json:"email"`
					Whatsapp string    `json:"whatsapp"`
				} `json:"client"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, "unpaid", resp.Data.Status)
		assert.Equal(t, float64(500000), resp.Data.TotalAmount)
		assert.Contains(t, resp.Data.InvoiceNumber, "INV-")
		assert.Contains(t, resp.Data.PaymentURL, "https://app.sandbox.midtrans.com/snap/v2/vtweb/")
		assert.Equal(t, "budi.santoso@example.com", resp.Data.Client.Email)

		// Direct DB verification
		var dbOrder models.Order
		err = tx.First(&dbOrder, "id = ?", resp.Data.OrderID).Error
		require.NoError(t, err)
		assert.Equal(t, models.OrderStatusUnpaid, dbOrder.Status)
		assert.Equal(t, resp.Data.Client.ID, dbOrder.ClientID)
	})
}

func TestPublicPaymentWebhook(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	// Create test package & client & unpaid order
	pkg := &models.Package{
		Name:           "Paket Diamond Webhook",
		Price:          1000000,
		FeaturesConfig: models.FeaturesConfig{"has_maps": true, "has_qr": true},
	}
	require.NoError(t, tx.Create(pkg).Error)

	client := &models.Client{
		Name:     "Siti Rahma",
		Email:    "siti@example.com",
		Whatsapp: "089876543210",
	}
	require.NoError(t, tx.Create(client).Error)

	order := &models.Order{
		InvoiceNumber: "INV-20261001-TEST01",
		ClientID:      client.ID,
		PackageID:     pkg.ID,
		TotalAmount:   1000000,
		Status:        models.OrderStatusUnpaid,
		PaymentURL:    "https://app.sandbox.midtrans.com/snap/v2/vtweb/test-token",
	}
	require.NoError(t, tx.Create(order).Error)

	t.Run("Webhook marks order as paid, generates tokens, and creates draft invitation", func(t *testing.T) {
		webhookPayload := models.PaymentWebhookRequest{
			OrderID:           order.InvoiceNumber,
			TransactionStatus: "settlement",
		}

		body, _ := json.Marshal(webhookPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/webhook/payment", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				OrderID      uuid.UUID  `json:"order_id"`
				Status       string     `json:"status"`
				FormToken    string     `json:"form_token"`
				ScannerToken string     `json:"scanner_token"`
				InvitationID *uuid.UUID `json:"invitation_id"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, "paid", resp.Data.Status)
		assert.Contains(t, resp.Data.FormToken, "form_")
		assert.Contains(t, resp.Data.ScannerToken, "scan_")
		require.NotNil(t, resp.Data.InvitationID)

		// Direct DB Verification on Order
		var updatedOrder models.Order
		err = tx.First(&updatedOrder, "id = ?", order.ID).Error
		require.NoError(t, err)
		assert.Equal(t, models.OrderStatusPaid, updatedOrder.Status)
		assert.NotEmpty(t, updatedOrder.FormToken)
		assert.NotEmpty(t, updatedOrder.ScannerToken)

		// Direct DB Verification on Invitations (auto-created 1 row draft)
		var invitation models.Invitation
		err = tx.First(&invitation, "order_id = ?", order.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "draft", invitation.Status)
		assert.Equal(t, client.ID, invitation.ClientID)
		assert.Equal(t, pkg.ID, invitation.PackageID)
		assert.Equal(t, *resp.Data.InvitationID, invitation.ID)
	})
}

// ====================================================
// 2. ADMIN CLIENTS CRUD, SEARCH, TRASH, & BULK ACTIONS
// ====================================================

func TestAdminClientsCRUD(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	t.Run("Create, Read, Update, Delete (Trash), Restore, Bulk Actions for Clients", func(t *testing.T) {
		// 1. Create client via Admin
		createPayload := models.CreateClientRequest{
			Name:     "Admin Client Test",
			Email:    "admin_client@example.com",
			Whatsapp: "081122334455",
		}
		createBody, _ := json.Marshal(createPayload)
		reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/clients", bytes.NewBuffer(createBody))
		reqCreate.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		router.ServeHTTP(wCreate, reqCreate)
		require.Equal(t, http.StatusCreated, wCreate.Code)

		var createResp struct {
			Data models.Client `json:"data"`
		}
		_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)
		clientID := createResp.Data.ID

		// 2. Get By ID
		reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/clients/"+clientID.String(), nil)
		wGet := httptest.NewRecorder()
		router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusOK, wGet.Code)

		// 3. Update client (PUT)
		newName := "Admin Client Updated"
		updatePayload := models.UpdateClientRequest{
			Name: &newName,
		}
		updateBody, _ := json.Marshal(updatePayload)
		reqUpdate, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/clients/"+clientID.String(), bytes.NewBuffer(updateBody))
		reqUpdate.Header.Set("Content-Type", "application/json")
		wUpdate := httptest.NewRecorder()
		router.ServeHTTP(wUpdate, reqUpdate)
		assert.Equal(t, http.StatusOK, wUpdate.Code)

		// 4. Soft Delete (moved to trash)
		reqDel, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/clients/"+clientID.String(), nil)
		wDel := httptest.NewRecorder()
		router.ServeHTTP(wDel, reqDel)
		assert.Equal(t, http.StatusOK, wDel.Code)

		// Verify query with is_trashed=true sees it
		reqTrash, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/clients?is_trashed=true", nil)
		wTrash := httptest.NewRecorder()
		router.ServeHTTP(wTrash, reqTrash)
		assert.Equal(t, http.StatusOK, wTrash.Code)

		// 5. Restore
		reqRestore, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/clients/"+clientID.String()+"/restore", nil)
		wRestore := httptest.NewRecorder()
		router.ServeHTTP(wRestore, reqRestore)
		assert.Equal(t, http.StatusOK, wRestore.Code)

		// 6. Bulk Delete & Bulk Restore
		bulkPayload := models.ClientBulkRequest{
			IDs: []uuid.UUID{clientID},
		}
		bulkBody, _ := json.Marshal(bulkPayload)

		reqBulkDel, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/clients/bulk-delete", bytes.NewBuffer(bulkBody))
		reqBulkDel.Header.Set("Content-Type", "application/json")
		wBulkDel := httptest.NewRecorder()
		router.ServeHTTP(wBulkDel, reqBulkDel)
		assert.Equal(t, http.StatusOK, wBulkDel.Code)

		reqBulkRes, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/clients/bulk-restore", bytes.NewBuffer(bulkBody))
		reqBulkRes.Header.Set("Content-Type", "application/json")
		wBulkRes := httptest.NewRecorder()
		router.ServeHTTP(wBulkRes, reqBulkRes)
		assert.Equal(t, http.StatusOK, wBulkRes.Code)
	})
}

// ====================================================
// 3. ADMIN ORDERS CRUD, FILTER, TRASH, & BULK ACTIONS
// ====================================================

func TestAdminOrdersCRUD(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	// Seed package & client
	pkg := &models.Package{
		Name:  "Paket Admin Orders",
		Price: 350000,
	}
	require.NoError(t, tx.Create(pkg).Error)

	client := &models.Client{
		Name:     "Admin Order Client",
		Email:    "order_client@example.com",
		Whatsapp: "082233445566",
	}
	require.NoError(t, tx.Create(client).Error)

	order := &models.Order{
		InvoiceNumber: "INV-20261001-ADMIN01",
		ClientID:      client.ID,
		PackageID:     pkg.ID,
		TotalAmount:   350000,
		Status:        models.OrderStatusUnpaid,
	}
	require.NoError(t, tx.Create(order).Error)

	t.Run("List with pagination/status, Get, Update, Delete, Restore, Bulk Actions", func(t *testing.T) {
		// 1. List with status filter
		reqList, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/orders?status=unpaid&page=1&limit=10", nil)
		wList := httptest.NewRecorder()
		router.ServeHTTP(wList, reqList)
		assert.Equal(t, http.StatusOK, wList.Code)

		// 2. Get Order by ID
		reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/orders/"+order.ID.String(), nil)
		wGet := httptest.NewRecorder()
		router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusOK, wGet.Code)

		// 3. Update Order (PUT)
		newStatus := models.OrderStatusPaid
		updatePayload := models.UpdateOrderRequest{
			Status: &newStatus,
		}
		updateBody, _ := json.Marshal(updatePayload)
		reqUpdate, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/orders/"+order.ID.String(), bytes.NewBuffer(updateBody))
		reqUpdate.Header.Set("Content-Type", "application/json")
		wUpdate := httptest.NewRecorder()
		router.ServeHTTP(wUpdate, reqUpdate)
		assert.Equal(t, http.StatusOK, wUpdate.Code)

		// 4. Soft Delete
		reqDel, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/orders/"+order.ID.String(), nil)
		wDel := httptest.NewRecorder()
		router.ServeHTTP(wDel, reqDel)
		assert.Equal(t, http.StatusOK, wDel.Code)

		// 5. Query trash via is_trashed=true
		reqTrash, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/orders?is_trashed=true", nil)
		wTrash := httptest.NewRecorder()
		router.ServeHTTP(wTrash, reqTrash)
		assert.Equal(t, http.StatusOK, wTrash.Code)

		// 6. Restore
		reqRestore, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/"+order.ID.String()+"/restore", nil)
		wRestore := httptest.NewRecorder()
		router.ServeHTTP(wRestore, reqRestore)
		assert.Equal(t, http.StatusOK, wRestore.Code)

		// 7. Bulk Delete & Bulk Restore
		bulkPayload := models.OrderBulkRequest{
			IDs: []uuid.UUID{order.ID},
		}
		bulkBody, _ := json.Marshal(bulkPayload)

		reqBulkDel, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/bulk-delete", bytes.NewBuffer(bulkBody))
		reqBulkDel.Header.Set("Content-Type", "application/json")
		wBulkDel := httptest.NewRecorder()
		router.ServeHTTP(wBulkDel, reqBulkDel)
		assert.Equal(t, http.StatusOK, wBulkDel.Code)

		reqBulkRes, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/orders/bulk-restore", bytes.NewBuffer(bulkBody))
		reqBulkRes.Header.Set("Content-Type", "application/json")
		wBulkRes := httptest.NewRecorder()
		router.ServeHTTP(wBulkRes, reqBulkRes)
		assert.Equal(t, http.StatusOK, wBulkRes.Code)
	})
}

func TestAdminClientsEagerLoadingOrdersAndPackages(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	// 1. Create package Platinum with has_qr == true
	pkg := &models.Package{
		Name:  "Platinum",
		Price: 750000,
		FeaturesConfig: models.FeaturesConfig{
			"has_qr":        true,
			"has_gallery":   true,
			"gallery_limit": 30,
		},
	}
	require.NoError(t, tx.Create(pkg).Error)

	// 2. Create client "Yosa"
	client := &models.Client{
		Name:     "Yosa",
		Email:    "yosa@example.com",
		Whatsapp: "081234567890",
	}
	require.NoError(t, tx.Create(client).Error)

	// 3. Create paid order for Yosa
	formToken := "abc"
	scannerToken := "def"
	order := &models.Order{
		InvoiceNumber: "INV-YOSA-001",
		ClientID:      client.ID,
		PackageID:     pkg.ID,
		TotalAmount:   750000,
		Status:        models.OrderStatusPaid,
		PaymentURL:    "https://app.midtrans.com/mock-yosa",
		FormToken:     &formToken,
		ScannerToken:  &scannerToken,
	}
	require.NoError(t, tx.Create(order).Error)

	// 4. Test GET /api/v1/admin/clients/:id
	t.Run("GET /api/v1/admin/clients/:id returns orders array with package and features_config", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/clients/"+client.ID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool          `json:"success"`
			Data    models.Client `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, client.ID, resp.Data.ID)
		assert.Equal(t, "Yosa", resp.Data.Name)
		require.Len(t, resp.Data.Orders, 1)

		o := resp.Data.Orders[0]
		assert.Equal(t, order.ID, o.ID)
		assert.Equal(t, models.OrderStatusPaid, o.Status)
		require.NotNil(t, o.FormToken)
		assert.Equal(t, "abc", *o.FormToken)
		require.NotNil(t, o.ScannerToken)
		assert.Equal(t, "def", *o.ScannerToken)

		require.NotNil(t, o.Package)
		assert.Equal(t, "Platinum", o.Package.Name)
		require.NotNil(t, o.Package.FeaturesConfig)
		assert.Equal(t, true, o.Package.FeaturesConfig["has_qr"])
	})

	// 5. Test GET /api/v1/admin/clients list
	t.Run("GET /api/v1/admin/clients list returns orders with preloaded package", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/clients?search=Yosa", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool            `json:"success"`
			Data    []models.Client `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.NotEmpty(t, resp.Data)
		found := false
		for _, c := range resp.Data {
			if c.ID == client.ID {
				found = true
				require.Len(t, c.Orders, 1)
				assert.Equal(t, "Platinum", c.Orders[0].Package.Name)
				assert.Equal(t, true, c.Orders[0].Package.FeaturesConfig["has_qr"])
				assert.Equal(t, "abc", *c.Orders[0].FormToken)
				assert.Equal(t, "def", *c.Orders[0].ScannerToken)
			}
		}
		assert.True(t, found, "Client Yosa should be present in response")
	})
}

