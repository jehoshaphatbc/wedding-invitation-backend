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

func TestClientPortal_AuthVerifyAndInvitation(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	router := setupTestRouter(tx)

	// Seed package, client, and order
	pkg := &models.Package{
		Name:  "Paket Platinum",
		Price: 1500000,
		FeaturesConfig: models.FeaturesConfig{
			"has_story":     true,
			"has_gallery":   true,
			"gallery_limit": 10,
			"has_gift":      true,
			"has_countdown": true,
			"has_maps":      true,
			"has_rsvp":      true,
			"has_qr":        true,
		},
	}
	require.NoError(t, tx.Create(pkg).Error)

	client := &models.Client{
		Name:     "Dimas & Anisa",
		Email:    "klien@example.com",
		Whatsapp: "081234567890",
	}
	require.NoError(t, tx.Create(client).Error)

	formToken := uuid.New().String()
	scannerToken := uuid.New().String()
	order := &models.Order{
		InvoiceNumber: "INV-20261001-0099",
		ClientID:      client.ID,
		PackageID:     pkg.ID,
		TotalAmount:   1500000,
		Status:        models.OrderStatusPaid,
		PaymentURL:    "https://example.com/payment",
		FormToken:     &formToken,
		ScannerToken:  &scannerToken,
	}
	require.NoError(t, tx.Create(order).Error)

	t.Run("AuthVerify: 401 when token is missing", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/client/auth-verify", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("AuthVerify: 401 when token is invalid", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/client/auth-verify?token=invalid-random-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("AuthVerify: 200 OK returns valid=true, token, client, order, package, and invitation=null when not yet created", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/client/auth-verify?token="+formToken, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                            `json:"success"`
			Message string                          `json:"message"`
			Data    models.ClientAuthVerifyResponse `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.True(t, resp.Data.Valid)
		assert.Equal(t, formToken, resp.Data.Token)
		assert.Equal(t, client.ID, resp.Data.Client.ID)
		assert.Equal(t, "Dimas & Anisa", resp.Data.Client.Name)
		assert.Equal(t, "klien@example.com", resp.Data.Client.Email)
		assert.Equal(t, "081234567890", resp.Data.Client.Whatsapp)
		assert.Equal(t, order.ID, resp.Data.Order.ID)
		assert.Equal(t, "INV-20261001-0099", resp.Data.Order.InvoiceNumber)
		assert.Equal(t, models.OrderStatusPaid, resp.Data.Order.Status)
		assert.Equal(t, "Paket Platinum", resp.Data.Package.Name)
		assert.Equal(t, true, resp.Data.Package.FeaturesConfig["has_qr"])
		assert.Equal(t, true, resp.Data.Package.FeaturesConfig["has_story"])
		assert.Nil(t, resp.Data.Invitation, "Invitation must be null before setup is saved")
	})

	t.Run("POST /api/v1/client/invitation: 401 when token is missing", func(t *testing.T) {
		payload := map[string]interface{}{
			"story": "Test story",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/client/invitation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST /api/v1/client/invitation: 200 OK with full setup payload and X-Client-Token header", func(t *testing.T) {
		payload := map[string]interface{}{
			"groom": map[string]interface{}{
				"full_name": "Muhammad Dimas Pratama, S.T.",
				"nickname":  "Dimas",
				"parents":   "Putra dari Bpk. Bambang & Ibu Sri",
				"instagram": "dimaspratama",
			},
			"bride": map[string]interface{}{
				"full_name": "Anisa Rahmawati, S.Ked.",
				"nickname":  "Anisa",
				"parents":   "Putri dari Bpk. Haryono & Ibu Endang",
				"instagram": "anisarahma",
			},
			"event": map[string]interface{}{
				"akad_date":              "2026-12-20",
				"akad_time":              "Pukul 08:00 - 10:00 WIB",
				"akad_time_start":        "08:00",
				"akad_time_end":          "10:00",
				"akad_is_until_end":      false,
				"akad_timezone":          "WIB",
				"reception_date":         "2026-12-20",
				"reception_time":         "Pukul 11:00 - 13:00 WIB",
				"reception_time_start":   "11:00",
				"reception_time_end":     "13:00",
				"reception_is_until_end": false,
				"reception_timezone":     "WIB",
				"is_same_location":       true,
				"venue_name":             "Grand Ballroom Hotel Mulia",
				"address":                "Jl. Asia Afrika No. 8, Jakarta Pusat",
				"maps_url":               "https://maps.app.goo.gl/test",
			},
			"theme": map[string]interface{}{
				"template_id":        "tpl-1",
				"template_component": "TemplateRomanticFloral",
				"primary_color":      "#B76E79",
			},
			"story": "Teks kisah cinta mempelai...",
			"gallery": []string{
				"https://drive.google.com/photo1",
				"https://drive.google.com/photo2",
			},
			"gift": map[string]interface{}{
				"bank_name":      "BCA",
				"account_number": "1234567890",
				"account_holder": "Muhammad Dimas Pratama",
			},
			"gifts": []map[string]interface{}{
				{
					"id":             "gift-1",
					"bank_name":      "BCA",
					"account_number": "1234567890",
					"account_holder": "Muhammad Dimas Pratama",
					"notes":          "Rekening Pria",
				},
				{
					"id":             "gift-2",
					"bank_name":      "Mandiri",
					"account_number": "9876543210",
					"account_holder": "Anisa Rahmawati",
					"notes":          "Rekening Wanita",
				},
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/client/invitation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Client-Token", formToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                   `json:"success"`
			Message string                 `json:"message"`
			Data    map[string]interface{} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, "Data undangan berhasil disimpan", resp.Message)
		assert.NotNil(t, resp.Data["id"])
		assert.NotNil(t, resp.Data["groom"])
		assert.NotNil(t, resp.Data["bride"])
		assert.NotNil(t, resp.Data["event"])
		assert.NotNil(t, resp.Data["theme"])
		assert.NotNil(t, resp.Data["gifts"])

		// Now verify auth-verify returns the saved invitation (not null anymore)
		reqVerify, _ := http.NewRequest(http.MethodGet, "/api/v1/client/auth-verify?token="+formToken, nil)
		wVerify := httptest.NewRecorder()
		router.ServeHTTP(wVerify, reqVerify)

		assert.Equal(t, http.StatusOK, wVerify.Code)
		var respVerify struct {
			Success bool                            `json:"success"`
			Data    models.ClientAuthVerifyResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(wVerify.Body.Bytes(), &respVerify))
		assert.NotNil(t, respVerify.Data.Invitation)
		assert.Equal(t, order.ID, respVerify.Data.Invitation.OrderID)
	})

	t.Run("PUT /api/v1/client/invitation: 200 OK updates existing invitation with Authorization Bearer", func(t *testing.T) {
		payload := map[string]interface{}{
			"story": "Kisah cinta yang diperbarui...",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/client/invitation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+formToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                   `json:"success"`
			Message string                 `json:"message"`
			Data    map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, "Data undangan berhasil disimpan", resp.Message)
		assert.Equal(t, "Kisah cinta yang diperbarui...", resp.Data["story"])
	})

	t.Run("POST /api/v1/invitation/setup: 200 OK alias works", func(t *testing.T) {
		payload := map[string]interface{}{
			"story": "Kisah cinta via setup endpoint",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/invitation/setup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+formToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                   `json:"success"`
			Message string                 `json:"message"`
			Data    map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, "Data undangan berhasil disimpan", resp.Message)
	})
}
