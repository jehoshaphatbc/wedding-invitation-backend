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

	// Seed package, client, order, and invitation
	pkg := &models.Package{
		Name:  "Paket Platinum Portal",
		Price: 1500000,
		FeaturesConfig: models.FeaturesConfig{
			"has_gallery":   true,
			"gallery_limit": 50,
			"has_qr":        true,
			"has_rsvp":      true,
		},
	}
	require.NoError(t, tx.Create(pkg).Error)

	client := &models.Client{
		Name:     "Rangga Wardana",
		Email:    "rangga.wardana@example.com",
		Whatsapp: "081234567899",
	}
	require.NoError(t, tx.Create(client).Error)

	formToken := uuid.New().String()
	scannerToken := uuid.New().String()
	order := &models.Order{
		InvoiceNumber: "INV-PORTAL-001",
		ClientID:      client.ID,
		PackageID:     pkg.ID,
		TotalAmount:   1500000,
		Status:        models.OrderStatusPaid,
		PaymentURL:    "https://example.com/payment",
		FormToken:     &formToken,
		ScannerToken:  &scannerToken,
	}
	require.NoError(t, tx.Create(order).Error)

	invitation := &models.Invitation{
		OrderID:   order.ID,
		ClientID:  client.ID,
		PackageID: pkg.ID,
		Title:     "Undangan Rangga & Cinta",
		Status:    "draft",
	}
	require.NoError(t, tx.Create(invitation).Error)

	t.Run("AuthVerify: 401 when token is missing", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/client/auth-verify", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("AuthVerify: 401 when token is invalid", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/client/auth-verify?token=invalid-random-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("AuthVerify: 200 OK with valid token via query param", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/client/auth-verify?token="+formToken, nil)
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
		assert.Equal(t, order.ID, resp.Data.Order.ID)
		assert.Equal(t, formToken, *resp.Data.Order.FormToken)
		assert.Equal(t, "Rangga Wardana", resp.Data.Client.Name)
		assert.Equal(t, "Paket Platinum Portal", resp.Data.Package.Name)
		assert.Equal(t, true, resp.Data.Package.FeaturesConfig["has_qr"])
		assert.Equal(t, true, resp.Data.Package.FeaturesConfig["has_gallery"])
		assert.NotNil(t, resp.Data.Invitation)
		assert.Equal(t, "Undangan Rangga & Cinta", resp.Data.Invitation.Title)
	})

	t.Run("AuthVerify: 200 OK with valid token via Authorization Bearer header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/client/auth-verify", nil)
		req.Header.Set("Authorization", "Bearer "+formToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateInvitation: 401 when Authorization header is missing", func(t *testing.T) {
		payload := models.UpdateClientInvitationRequest{
			GroomData: map[string]interface{}{"name": "Rangga", "father": "Bpk. Hendra"},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/api/client/invitation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UpdateInvitation: 401 when Bearer token is invalid", func(t *testing.T) {
		payload := models.UpdateClientInvitationRequest{
			GroomData: map[string]interface{}{"name": "Rangga", "father": "Bpk. Hendra"},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/api/client/invitation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer wrong-token-xyz")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UpdateInvitation: 200 OK updates groom, bride, events, story, gallery data", func(t *testing.T) {
		newTitle := "The Wedding of Rangga & Cinta"
		payload := models.UpdateClientInvitationRequest{
			Title: &newTitle,
			GroomData: map[string]interface{}{
				"name":      "Rangga Wardana",
				"nickname":  "Rangga",
				"father":    "Bpk. Hendra Wardana",
				"mother":    "Ibu Ratna",
				"instagram": "@rangga",
			},
			BrideData: map[string]interface{}{
				"name":      "Cinta Clarissa",
				"nickname":  "Cinta",
				"father":    "Bpk. Bram",
				"mother":    "Ibu Sinta",
				"instagram": "@cinta",
			},
			EventsData: []map[string]interface{}{
				{
					"name":     "Akad Nikah",
					"date":     "2026-12-12",
					"time":     "08:00 - 10:00",
					"location": "Masjid Raya Jakarta",
				},
				{
					"name":     "Resepsi",
					"date":     "2026-12-12",
					"time":     "11:00 - 14:00",
					"location": "Grand Ballroom Hotel Kempinski",
				},
			},
			StoryData: []map[string]interface{}{
				{
					"year":  "2020",
					"title": "Pertama Bertemu",
					"story": "Bertemu di kampus...",
				},
			},
			GalleryURLs: []string{
				"https://images.example.com/photo1.jpg",
				"https://images.example.com/photo2.jpg",
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/api/client/invitation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+formToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool              `json:"success"`
			Message string            `json:"message"`
			Data    models.Invitation `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, "The Wedding of Rangga & Cinta", resp.Data.Title)
		assert.Equal(t, "Rangga Wardana", resp.Data.GroomData["name"])
		assert.Equal(t, "Cinta Clarissa", resp.Data.BrideData["name"])
		assert.Len(t, resp.Data.GalleryURLs, 2)

		// Verify directly in DB
		var updatedInv models.Invitation
		require.NoError(t, tx.Where("order_id = ?", order.ID).First(&updatedInv).Error)
		assert.Equal(t, "The Wedding of Rangga & Cinta", updatedInv.Title)
		assert.Equal(t, "Rangga Wardana", updatedInv.GroomData["name"])
		assert.Equal(t, "Cinta Clarissa", updatedInv.BrideData["name"])
		assert.Len(t, updatedInv.GalleryURLs, 2)
	})
}
