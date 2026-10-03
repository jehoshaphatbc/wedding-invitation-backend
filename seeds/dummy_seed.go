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
	// STEP 0: TEMPLATES (2 Default Templates: Elegant White & Dark Rustic)
	// ------------------------------------------------------------------
	templatesData := []models.Template{
		{
			Name:          "Elegant White",
			Category:      "classic",
			NuxtComponent: "TemplateA",
			ThumbnailURL:  strPtr("https://images.unsplash.com/photo-1519741497674-611481863552"),
			IsActive:      true,
		},
		{
			Name:          "Dark Rustic",
			Category:      "rustic",
			NuxtComponent: "TemplateB",
			ThumbnailURL:  strPtr("https://images.unsplash.com/photo-1465495976277-4387d4b0b4c6"),
			IsActive:      true,
		},
	}

	for _, tmpl := range templatesData {
		var existing models.Template
		if err := tx.Where("name = ?", tmpl.Name).First(&existing).Error; err == nil {
			if existing.Category == "" {
				existing.Category = tmpl.Category
				_ = tx.Save(&existing)
			}
			continue
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&tmpl).Error; err != nil {
			return nil, fmt.Errorf("failed to seed template %s: %w", tmpl.Name, err)
		}
	}

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
	// STEP 4: INVITATIONS (Scenario 1: Completed Setup, Scenario 2: Draft/Nil)
	// ------------------------------------------------------------------
	slug1 := "budi-ani"
	slug2 := "siti-rizky"

	var targetInvitationID uuid.UUID

	for i, po := range paidOrders {
		switch i {
		case 0:
			// Skenario 1: Order yang benar-benar sudah mengisi setup (lengkap & published)
			slugCandidate := slug1
			var existingSlug models.Invitation
			if err := tx.Where("slug = ? AND order_id != ?", slugCandidate, po.ID).First(&existingSlug).Error; err == nil {
				slugCandidate = fmt.Sprintf("%s-auto", slugCandidate)
			}

			groomData := map[string]interface{}{
				"full_name": "Budi Pratama, S.Kom.",
				"nickname":  "Budi",
				"parents":   "Putra dari Bpk. Santoso & Ibu Ratna",
				"instagram": "budipratama",
			}
			brideData := map[string]interface{}{
				"full_name": "Ani Wijaya, S.E.",
				"nickname":  "Ani",
				"parents":   "Putri dari Bpk. Bambang & Ibu Siti",
				"instagram": "aniwijaya",
			}
			eventData := map[string]interface{}{
				"akad_date":            "2026-12-25",
				"akad_time":            "Pukul 08:00 - 10:00 WIB",
				"akad_time_start":      "08:00",
				"akad_time_end":        "10:00",
				"akad_timezone":        "WIB",
				"reception_date":       "2026-12-25",
				"reception_time":       "Pukul 11:00 - 13:00 WIB",
				"reception_time_start": "11:00",
				"reception_time_end":   "13:00",
				"reception_timezone":   "WIB",
				"is_same_location":     true,
				"venue_name":           "Grand Ballroom Hotel Mulia",
				"address":              "Jl. Asia Afrika No. 8, Jakarta Pusat",
				"maps_url":             "https://maps.app.goo.gl/dummy",
			}
			themeData := map[string]interface{}{
				"template_id":        "tpl-1",
				"template_component": "TemplateRomanticFloral",
				"primary_color":      "#B76E79",
			}
			storyData := "Perjalanan cinta kami dimulai di bangku kuliah hingga akhirnya memutuskan melangkah ke jenjang pernikahan."
			galleryData := []string{
				"https://images.unsplash.com/photo-1519741497674-611481863552",
				"https://images.unsplash.com/photo-1511285560929-80b456fea0bc",
			}
			giftsData := []map[string]interface{}{
				{
					"id":             "gift-1",
					"bank_name":      "BCA",
					"account_number": "1234567890",
					"account_holder": "Budi Pratama",
					"notes":          "Rekening Utama",
				},
			}

			inv := models.Invitation{
				OrderID:   po.ID,
				ClientID:  po.ClientID,
				PackageID: po.PackageID,
				Title:     "The Wedding of Budi & Ani",
				Slug:      &slugCandidate,
				Status:    "published",
				Groom:     groomData,
				Bride:     brideData,
				Event:     eventData,
				Theme:     themeData,
				Story:     storyData,
				Gallery:   galleryData,
				Gifts:     giftsData,
			}

			var existingInv models.Invitation
			if err := tx.Where("order_id = ?", po.ID).First(&existingInv).Error; err == nil {
				existingInv.Title = inv.Title
				existingInv.Slug = inv.Slug
				existingInv.Status = inv.Status
				existingInv.Groom = inv.Groom
				existingInv.Bride = inv.Bride
				existingInv.Event = inv.Event
				existingInv.Theme = inv.Theme
				existingInv.Story = inv.Story
				existingInv.Gallery = inv.Gallery
				existingInv.Gifts = inv.Gifts
				if err := tx.Save(&existingInv).Error; err != nil {
					return nil, fmt.Errorf("failed to update invitation %s: %w", inv.Title, err)
				}
				targetInvitationID = existingInv.ID
			} else {
				if err := tx.Create(&inv).Error; err != nil {
					return nil, fmt.Errorf("failed to seed invitation %s: %w", inv.Title, err)
				}
				targetInvitationID = inv.ID
			}

		case 1:
			// Skenario 2a: Order baru yang belum mengisi formulir (status "draft", data kosong)
			slugCandidate := slug2
			var existingSlug models.Invitation
			if err := tx.Where("slug = ? AND order_id != ?", slugCandidate, po.ID).First(&existingSlug).Error; err == nil {
				slugCandidate = fmt.Sprintf("%s-auto", slugCandidate)
			}

			inv := models.Invitation{
				OrderID:   po.ID,
				ClientID:  po.ClientID,
				PackageID: po.PackageID,
				Title:     "The Wedding of Siti & Rizky",
				Slug:      &slugCandidate,
				Status:    "draft",
			}

			var existingInv models.Invitation
			if err := tx.Where("order_id = ?", po.ID).First(&existingInv).Error; err == nil {
				existingInv.Title = inv.Title
				existingInv.Slug = inv.Slug
				existingInv.Status = "draft"
				existingInv.Groom = nil
				existingInv.Bride = nil
				existingInv.Event = nil
				existingInv.Theme = nil
				existingInv.Story = nil
				existingInv.Gallery = nil
				existingInv.Gift = nil
				existingInv.Gifts = nil
				if err := tx.Save(&existingInv).Error; err != nil {
					return nil, fmt.Errorf("failed to update draft invitation: %w", err)
				}
			} else {
				if err := tx.Create(&inv).Error; err != nil {
					return nil, fmt.Errorf("failed to seed draft invitation: %w", err)
				}
			}

		case 2:
			// Skenario 2b: Order baru yang belum ada record invitation sama sekali (invitation: nil)
			if err := tx.Where("order_id = ?", po.ID).Delete(&models.Invitation{}).Error; err != nil {
				return nil, fmt.Errorf("failed to clean up invitation for order 3: %w", err)
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
		"templates_seeded":   len(templatesData),
		"packages_seeded":    len(packages),
		"clients_seeded":     len(clients),
		"orders_seeded":      len(ordersData),
		"invitations_seeded": 2,
		"guests_seeded":      10,
	}

	return result, nil
}
