package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
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
	var packagesCount, clientsCount, ordersCount, invitationsCount, guestsCount int

	err := h.db.Transaction(func(tx *gorm.DB) error {
		// ------------------------------------------------------------------
		// STEP 1: PACKAGES (3 Data: Silver, Gold, Platinum with JSONB)
		// ------------------------------------------------------------------
		packagesData := []models.Package{
			{
				Name:  "Paket Silver",
				Price: 250000,
				FeaturesConfig: models.FeaturesConfig{
					"has_countdown": true,
					"has_maps":      true,
					"has_rsvp":      true,
					"has_gift":      true,
					"has_gallery":   true,
					"gallery_limit": 5,
					"has_story":     false,
					"has_video":     false,
					"has_qr":        false,
				},
			},
			{
				Name:  "Paket Gold",
				Price: 450000,
				FeaturesConfig: models.FeaturesConfig{
					"has_countdown": true,
					"has_maps":      true,
					"has_rsvp":      true,
					"has_gift":      true,
					"has_gallery":   true,
					"gallery_limit": 15,
					"has_story":     true,
					"has_video":     true,
					"has_qr":        false,
				},
			},
			{
				Name:  "Paket Platinum",
				Price: 750000,
				FeaturesConfig: models.FeaturesConfig{
					"has_countdown": true,
					"has_maps":      true,
					"has_rsvp":      true,
					"has_gift":      true,
					"has_gallery":   true,
					"gallery_limit": 30,
					"has_story":     true,
					"has_video":     true,
					"has_qr":        true,
				},
			},
		}

		packages := make([]models.Package, len(packagesData))
		for i, pkg := range packagesData {
			var persistedPkg models.Package
			if err := tx.Where("name = ?", pkg.Name).First(&persistedPkg).Error; err == nil {
				// Package already exists, reuse it
				packages[i] = persistedPkg
				continue
			}

			// ON CONFLICT DO NOTHING to avoid duplicate crashes
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&pkg).Error; err != nil {
				return fmt.Errorf("failed to seed packages: %w", err)
			}

			// Retrieve by name to get valid ID
			if err := tx.Where("name = ?", pkg.Name).First(&persistedPkg).Error; err != nil {
				return fmt.Errorf("failed to fetch package ID: %w", err)
			}
			packages[i] = persistedPkg
		}
		packagesCount = len(packages)

		// ------------------------------------------------------------------
		// STEP 2: CLIENTS (3 Data: Indonesian Local Names & Emails)
		// ------------------------------------------------------------------
		clientsData := []models.Client{
			{
				Name:     "Budi Pratama",
				Email:    "budi.pratama@gmail.com",
				Whatsapp: "081298765432",
			},
			{
				Name:     "Siti Nurhaliza",
				Email:    "siti.nurhaliza@gmail.com",
				Whatsapp: "085712345678",
			},
			{
				Name:     "Dimas Setiawan",
				Email:    "dimas.setiawan@gmail.com",
				Whatsapp: "087811223344",
			},
		}

		clients := make([]models.Client, len(clientsData))
		for i, client := range clientsData {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&client).Error; err != nil {
				return fmt.Errorf("failed to seed clients: %w", err)
			}

			var persistedClient models.Client
			if err := tx.Where("email = ?", client.Email).First(&persistedClient).Error; err != nil {
				return fmt.Errorf("failed to fetch client ID: %w", err)
			}
			clients[i] = persistedClient
		}
		clientsCount = len(clients)

		// ------------------------------------------------------------------
		// STEP 3: ORDERS (5 Transactions: 3 Paid with Tokens, 2 Unpaid)
		// ------------------------------------------------------------------
		ordersData := []struct {
			InvoiceNumber string
			ClientIdx     int
			PackageIdx    int
			Status        models.OrderStatus
		}{
			{"INV-SEED-001", 0, 1, models.OrderStatusPaid},   // Budi, Gold, Paid
			{"INV-SEED-002", 1, 2, models.OrderStatusPaid},   // Siti, Platinum, Paid
			{"INV-SEED-003", 2, 0, models.OrderStatusPaid},   // Dimas, Silver, Paid
			{"INV-SEED-004", 0, 0, models.OrderStatusUnpaid}, // Budi, Silver, Unpaid
			{"INV-SEED-005", 1, 1, models.OrderStatusUnpaid}, // Siti, Gold, Unpaid
		}

		paidOrders := make([]models.Order, 0)
		for _, od := range ordersData {
			client := clients[od.ClientIdx]
			pkg := packages[od.PackageIdx]

			var formToken, scannerToken string
			if od.Status == models.OrderStatusPaid {
				formToken = "form_" + uuid.New().String()
				scannerToken = "scan_" + uuid.New().String()
			}

			order := models.Order{
				InvoiceNumber: od.InvoiceNumber,
				ClientID:      client.ID,
				PackageID:     pkg.ID,
				TotalAmount:   pkg.Price,
				Status:        od.Status,
				PaymentURL:    fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", uuid.New().String()),
				FormToken:     formToken,
				ScannerToken:  scannerToken,
			}

			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&order).Error; err != nil {
				return fmt.Errorf("failed to seed order: %w", err)
			}

			var persistedOrder models.Order
			if err := tx.Where("invoice_number = ?", od.InvoiceNumber).First(&persistedOrder).Error; err != nil {
				return fmt.Errorf("failed to fetch order ID: %w", err)
			}

			if persistedOrder.Status == models.OrderStatusPaid {
				paidOrders = append(paidOrders, persistedOrder)
			}
		}
		ordersCount = len(ordersData)

		// ------------------------------------------------------------------
		// STEP 4: INVITATIONS (3 Data related to Paid Orders with Slugs)
		// ------------------------------------------------------------------
		slug1 := "budi-ani"
		slug2 := "siti-rizky"
		slug3 := "dimas-putri"

		invitationsData := []struct {
			Title string
			Slug  *string
		}{
			{"The Wedding of Budi & Ani", &slug1},
			{"The Wedding of Siti & Rizky", &slug2},
			{"The Wedding of Dimas & Putri", &slug3},
		}

		var targetInvitationID uuid.UUID

		for i, po := range paidOrders {
			if i >= len(invitationsData) {
				break
			}
			invData := invitationsData[i]
			slugCandidate := *invData.Slug
			var existingSlug models.Invitation
			if err := tx.Where("slug = ? AND order_id != ?", slugCandidate, po.ID).First(&existingSlug).Error; err == nil {
				slugCandidate = fmt.Sprintf("%s-http", slugCandidate)
			}

			inv := models.Invitation{
				OrderID:   po.ID,
				ClientID:  po.ClientID,
				PackageID: po.PackageID,
				Title:     invData.Title,
				Slug:      &slugCandidate,
				Status:    "published",
			}

			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&inv).Error; err != nil {
				return fmt.Errorf("failed to seed invitation: %w", err)
			}

			var persistedInv models.Invitation
			if err := tx.Where("order_id = ?", po.ID).First(&persistedInv).Error; err == nil {
				if i == 0 {
					targetInvitationID = persistedInv.ID
				}
			}
		}
		invitationsCount = len(invitationsData)

		// ------------------------------------------------------------------
		// STEP 5: GUESTS (10 Guests for Invitation 1 with QR and RSVP)
		// ------------------------------------------------------------------
		if targetInvitationID != uuid.Nil {
			guestsData := []struct {
				Name             string
				Phone            string
				RSVPStatus       string
				ActualAttendance bool
			}{
				{"Hendra Gunawan", "081211112222", "hadir", true},
				{"Eko Prasetyo", "081222223333", "hadir", true},
				{"Bambang Suryono", "081233334444", "hadir", false},
				{"Agus Wijaya", "081244445555", "hadir", false},
				{"Rina Nose", "081322223333", "tidak_hadir", false},
				{"Doni Kusuma", "081333334444", "tidak_hadir", false},
				{"Aditya Pratama", "081377778888", "pending", false},
				{"Bagus Triadi", "081388889999", "pending", false},
				{"Citra Kirana", "081399990000", "pending", false},
				{"Danang Sutrisno", "081411112222", "pending", false},
			}

			now := time.Now().Add(-1 * time.Hour)

			for i, gd := range guestsData {
				qrToken := fmt.Sprintf("QR-SEED-%03d-%s", i+1, targetInvitationID.String()[:8])
				phone := gd.Phone
				var attTime *time.Time
				if gd.ActualAttendance {
					tTime := now.Add(time.Duration(i*10) * time.Minute)
					attTime = &tTime
				}

				guest := models.Guest{
					InvitationID:     targetInvitationID,
					Name:             gd.Name,
					Phone:            &phone,
					Pax:              1,
					QRToken:          qrToken,
					RSVPStatus:       gd.RSVPStatus,
					ActualAttendance: gd.ActualAttendance,
					AttendanceTime:   attTime,
				}

				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&guest).Error; err != nil {
					return fmt.Errorf("failed to seed guest: %w", err)
				}
			}
			guestsCount = len(guestsData)
		}

		return nil
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
			"packages_seeded":    packagesCount,
			"clients_seeded":     clientsCount,
			"orders_seeded":      ordersCount,
			"invitations_seeded": invitationsCount,
			"guests_seeded":      guestsCount,
		},
	})
}
