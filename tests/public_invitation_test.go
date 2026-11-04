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
	"github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

func TestPublicInvitationEndpoints(t *testing.T) {
	tx, cleanup := setupTx(t)
	defer cleanup()

	// Seed database with dummy seed
	_, err := seeds.SeedDummyData(tx)
	require.NoError(t, err)

	router := setupTestRouter(tx)

	// Retrieve seeded invitation
	var invitation models.Invitation
	err = tx.Where("slug = ?", "mulyo-cempaka").First(&invitation).Error
	require.NoError(t, err)

	t.Run("GET /api/v1/invitations/slug/:slug - Success with Exact Data Schema", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/invitations/slug/mulyo-cempaka", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Status string                       `json:"status"`
			Data   *models.PublicInvitationData `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "success", resp.Status)
		require.NotNil(t, resp.Data)

		// 1. Root-level invitation metadata
		assert.Equal(t, invitation.ID, resp.Data.ID)
		assert.Equal(t, "mulyo-cempaka", resp.Data.Slug)
		assert.NotNil(t, resp.Data.TemplateID)

		// 2. Template verification
		assert.Equal(t, "Scarlet Sketch", resp.Data.Template.Name)
		assert.Equal(t, "TemplateScarletSketch", resp.Data.Template.NuxtComponent)
		assert.Equal(t, "scarlet-sketch", resp.Data.Template.Slug)

		// 3. Features config verification
		assert.True(t, resp.Data.FeaturesConfig["has_countdown"].(bool))
		assert.True(t, resp.Data.FeaturesConfig["has_gallery"].(bool))
		assert.True(t, resp.Data.FeaturesConfig["has_gift"].(bool))
		assert.True(t, resp.Data.FeaturesConfig["has_rsvp"].(bool))

		// 4. Invitation content verification
		content := resp.Data.Invitation
		require.NotNil(t, content)

		assert.Equal(t, "https://storage.domain.com/audio/God-Bless-The-Broken-Road.mp3", content["audio_url"])
		assert.NotEmpty(t, content["verse_text"])
		assert.Equal(t, "KOLOSE 3:14", content["verse_ref"])
		assert.NotEmpty(t, content["countdown_target_date"])

		// Groom details
		groom := content["groom"].(map[string]any)
		assert.Equal(t, "Mulyo Aryawijaya", groom["name"])
		assert.Equal(t, "Mulyo", groom["nickname"])
		assert.Equal(t, "Bapak Arya", groom["father"])
		assert.Equal(t, "Ibu Wijaya", groom["mother"])
		assert.Equal(t, "mulyoarya", groom["instagram"])

		// Bride details
		bride := content["bride"].(map[string]any)
		assert.Equal(t, "Putu Cempaka Giriyanti", bride["name"])
		assert.Equal(t, "Cempaka", bride["nickname"])
		assert.Equal(t, "Bapak Giri", bride["father"])
		assert.Equal(t, "Ibu Yanti", bride["mother"])
		assert.Equal(t, "cempakagiri", bride["instagram"])

		// Events
		events := content["events"].(map[string]any)
		require.NotNil(t, events)
		require.Contains(t, events, "holy_matrimony")
		require.Contains(t, events, "reception")

		holyMatrimony := events["holy_matrimony"].(map[string]any)
		assert.Equal(t, "Holy Matrimony", holyMatrimony["title"])
		assert.Equal(t, "Gereja Katedral Santo Roh Kudus", holyMatrimony["location"])

		reception := events["reception"].(map[string]any)
		assert.Equal(t, "Wedding Reception", reception["title"])
		assert.Equal(t, "The Glass House Bali", reception["location"])

		// Gift
		gift := content["gift"].(map[string]any)
		require.NotNil(t, gift)
		assert.Equal(t, "BCA", gift["bank_name"])
		assert.Equal(t, "127326134", gift["account_number"])
		assert.Equal(t, "Putu Cempaka Giriyanti Aryawijaya", gift["account_name"])

		// Gallery
		assert.NotEmpty(t, content["gallery"])
	})

	t.Run("GET /api/v1/invitations/slug/non-existent - 404 Not Found", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/invitations/slug/not-found-slug", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("POST /api/v1/invitations/:id/rsvp - Success RSVP and Wish Submission", func(t *testing.T) {
		payload := map[string]any{
			"guest_name":        "Test Guest",
			"attendance_status": "hadir",
			"number_of_guests":  2,
			"message":           "Selamat atas pernikahannya! Bahagia selalu selamanya.",
		}
		bodyBytes, _ := json.Marshal(payload)

		url := fmt.Sprintf("/api/v1/invitations/%s/rsvp", invitation.ID.String())
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp["status"])

		// Verify wish is stored in database
		var count int64
		tx.Model(&models.Wish{}).
			Where("invitation_id = ? AND guest_name = ?", invitation.ID, "Test Guest").
			Count(&count)
		assert.Equal(t, int64(1), count)

		// Verify guest is also recorded/updated in guests table
		var guest models.Guest
		err = tx.Where("invitation_id = ? AND name = ?", invitation.ID, "Test Guest").First(&guest).Error
		require.NoError(t, err)
		assert.Equal(t, "hadir", guest.RSVPStatus)
		assert.Equal(t, 2, guest.Pax)
		assert.NotEmpty(t, guest.QRToken)
	})

	t.Run("GET /api/v1/invitations/:id/wishes - Retrieve Wishes List", func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/invitations/%s/wishes", invitation.ID.String())
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Status string        `json:"status"`
			Data   []models.Wish `json:"data"`
			Total  int64         `json:"total"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "success", resp.Status)
		assert.GreaterOrEqual(t, resp.Total, int64(3))
		assert.NotEmpty(t, resp.Data)

		// Ensure Arthur and Nanik's wishes are in the list
		var foundArthur bool
		for _, w := range resp.Data {
			if w.GuestName == "Arthur" {
				foundArthur = true
				break
			}
		}
		assert.True(t, foundArthur, "Arthur's wish should be returned")
	})

	t.Run("POST /api/v1/invitations/:id/rsvp - Invalid Invitation ID", func(t *testing.T) {
		payload := map[string]any{
			"guest_name": "Ghost",
			"message":    "Hello",
		}
		bodyBytes, _ := json.Marshal(payload)

		url := fmt.Sprintf("/api/v1/invitations/%s/rsvp", uuid.New().String())
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
