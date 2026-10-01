package seeds

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

func strPtr(s string) *string {
	return &s
}

// SeedDummyData populates packages, clients, orders, invitations, and guests.
// It uses idempotent queries and ON CONFLICT DO NOTHING so that it can be safely run
// automatically on startup/migration or via the HTTP seeder endpoint.
func SeedDummyData(tx *gorm.DB) (map[string]int, error) {
	// ------------------------------------------------------------------
	// STEP 1: PACKAGES (3 Data: Silver, Gold, Platinum with features_config)
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

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&pkg).Error; err != nil {
			return nil, fmt.Errorf("failed to seed package %s: %w", pkg.Name, err)
		}

		if err := tx.Where("name = ?", pkg.Name).First(&persistedPkg).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch package ID for %s: %w", pkg.Name, err)
		}
		packages[i] = persistedPkg
	}

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
	for i, cl := range clientsData {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&cl).Error; err != nil {
			return nil, fmt.Errorf("failed to seed clients: %w", err)
		}

		var persistedClient models.Client
		if err := tx.Where("email = ?", cl.Email).First(&persistedClient).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch client by email %s: %w", cl.Email, err)
		}
		clients[i] = persistedClient
	}

	// ------------------------------------------------------------------
	// STEP 3: ORDERS (5 Data: 3 Paid with tokens, 2 Unpaid)
	// ------------------------------------------------------------------
	ordersData := []struct {
		InvoiceNumber string
		ClientIdx     int
		PackageIdx    int
		Amount        float64
		Status        models.OrderStatus
		IsPaid        bool
	}{
		{"INV-SEED-001", 0, 0, 250000, models.OrderStatusPaid, true},
		{"INV-SEED-002", 1, 1, 450000, models.OrderStatusPaid, true},
		{"INV-SEED-003", 2, 2, 750000, models.OrderStatusPaid, true},
		{"INV-SEED-004", 0, 1, 450000, models.OrderStatusUnpaid, false},
		{"INV-SEED-005", 1, 0, 250000, models.OrderStatusUnpaid, false},
	}

	var paidOrders []models.Order
	for _, o := range ordersData {
		var formToken, scannerToken *string
		paymentURL := fmt.Sprintf("https://app.midtrans.com/snap/v2/vtweb/mock-%s", o.InvoiceNumber)

		if o.IsPaid {
			ft := "form_" + uuid.New().String()
			formToken = &ft

			// Check related package JSONB features_config -> has_qr
			relPkg := packages[o.PackageIdx]
			hasQR := false
			if relPkg.FeaturesConfig != nil {
				if val, ok := relPkg.FeaturesConfig["has_qr"]; ok {
					if b, isBool := val.(bool); isBool && b {
						hasQR = true
					}
				}
			}

			if hasQR {
				st := "scan_" + uuid.New().String()
				scannerToken = &st
			} else {
				scannerToken = nil
			}
		}

		var persistedOrder models.Order
		if err := tx.Where("invoice_number = ?", o.InvoiceNumber).First(&persistedOrder).Error; err == nil {
			persistedOrder.FormToken = formToken
			persistedOrder.ScannerToken = scannerToken
			persistedOrder.Status = o.Status
			persistedOrder.PackageID = packages[o.PackageIdx].ID
			persistedOrder.ClientID = clients[o.ClientIdx].ID
			persistedOrder.TotalAmount = o.Amount
			persistedOrder.PaymentURL = paymentURL
			if err := tx.Save(&persistedOrder).Error; err != nil {
				return nil, fmt.Errorf("failed to update order %s: %w", o.InvoiceNumber, err)
			}
		} else {
			order := models.Order{
				InvoiceNumber: o.InvoiceNumber,
				ClientID:      clients[o.ClientIdx].ID,
				PackageID:     packages[o.PackageIdx].ID,
				TotalAmount:   o.Amount,
				Status:        o.Status,
				PaymentURL:    paymentURL,
				FormToken:     formToken,
				ScannerToken:  scannerToken,
			}

			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&order).Error; err != nil {
				return nil, fmt.Errorf("failed to seed order %s: %w", o.InvoiceNumber, err)
			}

			if err := tx.Where("invoice_number = ?", o.InvoiceNumber).First(&persistedOrder).Error; err != nil {
				return nil, fmt.Errorf("failed to fetch order %s: %w", o.InvoiceNumber, err)
			}
		}

		if persistedOrder.Status == models.OrderStatusPaid {
			paidOrders = append(paidOrders, persistedOrder)
		}
	}

	// ------------------------------------------------------------------
	// STEP 4: INVITATIONS (3 Data linked to Paid Orders)
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
			slugCandidate = fmt.Sprintf("%s-auto", slugCandidate)
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
			return nil, fmt.Errorf("failed to seed invitation: %w", err)
		}

		var persistedInv models.Invitation
		if err := tx.Where("order_id = ?", po.ID).First(&persistedInv).Error; err == nil {
			if i == 0 {
				targetInvitationID = persistedInv.ID
			}
		}
	}

	// ------------------------------------------------------------------
	// STEP 5: GUESTS (10 Data for first Invitation)
	// ------------------------------------------------------------------
	if targetInvitationID != uuid.Nil {
		now := time.Now()
		t1 := now.Add(-1 * time.Hour)
		t2 := now.Add(-50 * time.Minute)

		guestsData := []models.Guest{
			{Name: "Hendra Gunawan", Phone: strPtr("081211112222"), Pax: 1, RSVPStatus: "hadir", ActualAttendance: true, AttendanceTime: &t1},
			{Name: "Eko Prasetyo", Phone: strPtr("081222223333"), Pax: 1, RSVPStatus: "hadir", ActualAttendance: true, AttendanceTime: &t2},
			{Name: "Bambang Suryono", Phone: strPtr("081233334444"), Pax: 1, RSVPStatus: "hadir", ActualAttendance: false},
			{Name: "Agus Wijaya", Phone: strPtr("081244445555"), Pax: 1, RSVPStatus: "hadir", ActualAttendance: false},
			{Name: "Rina Nose", Phone: strPtr("081322223333"), Pax: 1, RSVPStatus: "tidak_hadir", ActualAttendance: false},
			{Name: "Doni Kusuma", Phone: strPtr("081333334444"), Pax: 1, RSVPStatus: "tidak_hadir", ActualAttendance: false},
			{Name: "Aditya Pratama", Phone: strPtr("081377778888"), Pax: 1, RSVPStatus: "pending", ActualAttendance: false},
			{Name: "Bagus Triadi", Phone: strPtr("081388889999"), Pax: 1, RSVPStatus: "pending", ActualAttendance: false},
			{Name: "Citra Kirana", Phone: strPtr("081399990000"), Pax: 1, RSVPStatus: "pending", ActualAttendance: false},
			{Name: "Danang Sutrisno", Phone: strPtr("081411112222"), Pax: 1, RSVPStatus: "pending", ActualAttendance: false},
		}

		for idx, g := range guestsData {
			g.InvitationID = targetInvitationID
			g.QRToken = fmt.Sprintf("QR-SEED-%03d-%s", idx+1, targetInvitationID.String()[:8])

			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&g).Error; err != nil {
				return nil, fmt.Errorf("failed to seed guest: %w", err)
			}
		}
	}

	result := map[string]int{
		"packages_seeded":    len(packages),
		"clients_seeded":     len(clients),
		"orders_seeded":      len(ordersData),
		"invitations_seeded": len(invitationsData),
		"guests_seeded":      10,
	}

	return result, nil
}
