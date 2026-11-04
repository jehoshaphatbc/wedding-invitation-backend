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
	// STEP 0: TEMPLATES (Noir & Scarlet Sketch, check migration from Harsava)
	// ------------------------------------------------------------------
	// Check migration / renaming: If "Harsava" exists, update to "Noir"
	tx.Model(&models.Template{}).
		Where("name ILIKE ? OR nuxt_component ILIKE ?", "%harsava%", "%harsava%").
		Updates(map[string]interface{}{
			"name":           "Noir",
			"slug":           "noir",
			"nuxt_component": "TemplateNoir",
			"category":       "Modern Elegant",
		})

	descNoir := "Tema Luxury Dark Minimalist. Sticky desktop preview di kiri, scrollable content di kanan, responsive full width di tablet/mobile, interactive envelope cover modal."
	descScarlet := "Tema Hand-drawn Sketch Wedding. Dominasi warna Scarlet Maroon (#9E0F00) dan Paper Background (#F5F2F0), tipografi Schoolbell, audio player timeline, amplop popup modal awal."

	templatesData := []models.Template{
		{
			Name:          "Noir",
			Slug:          "noir",
			Category:      "Modern Elegant",
			NuxtComponent: "TemplateNoir",
			ThumbnailURL:  strPtr("/images/templates/noir-thumbnail.jpg"),
			Description:   &descNoir,
			DefaultFeaturesConfig: models.FeaturesConfig{
				"has_countdown": true,
				"has_gallery":   true,
				"has_story":     true,
				"has_gift":      true,
				"has_rsvp":      true,
				"has_video":     false,
				"has_maps":      true,
				"has_qr":        true,
				"has_audio":     true,
			},
			IsActive: true,
		},
		{
			Name:          "Scarlet Sketch",
			Slug:          "scarlet-sketch",
			Category:      "Vintage Sketch",
			NuxtComponent: "TemplateScarletSketch",
			ThumbnailURL:  strPtr("/images/templates/scarlet-sketch-thumbnail.jpg"),
			Description:   &descScarlet,
			DefaultFeaturesConfig: models.FeaturesConfig{
				"has_countdown": true,
				"has_gallery":   true,
				"has_story":     false,
				"has_gift":      true,
				"has_rsvp":      true,
				"has_video":     false,
				"has_maps":      true,
				"has_qr":        false,
				"has_audio":     true,
			},
			IsActive: true,
		},
	}

	persistedTemplates := make(map[string]models.Template)
	for _, tmpl := range templatesData {
		var existing models.Template
		if err := tx.Where("name = ? OR slug = ? OR nuxt_component = ?", tmpl.Name, tmpl.Slug, tmpl.NuxtComponent).First(&existing).Error; err == nil {
			existing.Name = tmpl.Name
			existing.Slug = tmpl.Slug
			existing.Category = tmpl.Category
			existing.NuxtComponent = tmpl.NuxtComponent
			existing.ThumbnailURL = tmpl.ThumbnailURL
			existing.Description = tmpl.Description
			existing.DefaultFeaturesConfig = tmpl.DefaultFeaturesConfig
			existing.IsActive = tmpl.IsActive
			_ = tx.Save(&existing)
			persistedTemplates[tmpl.Slug] = existing
			continue
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&tmpl).Error; err != nil {
			return nil, fmt.Errorf("failed to seed template %s: %w", tmpl.Name, err)
		}
		var created models.Template
		if err := tx.Where("slug = ?", tmpl.Slug).First(&created).Error; err == nil {
			persistedTemplates[tmpl.Slug] = created
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
	slug1 := "mulyo-cempaka"
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

			audioURL := "https://storage.domain.com/audio/God-Bless-The-Broken-Road.mp3"
			verseText := "Dan di atas semuanya itu: kenakanlah kasih, sebagai pengikat yang mempersatukan dan menyempurnakan."
			verseRef := "KOLOSE 3:14"
			countdownDate := "2026-05-18T09:00:00+08:00"

			groomData := map[string]interface{}{
				"name":      "Mulyo Aryawijaya",
				"full_name": "Mulyo Aryawijaya",
				"nickname":  "Mulyo",
				"father":    "Bapak Arya",
				"mother":    "Ibu Wijaya",
				"parents":   "Putra dari Bapak Arya & Ibu Wijaya",
				"instagram": "mulyoarya",
				"photo_url": "https://storage.domain.com/photos/groom.jpg",
			}
			brideData := map[string]interface{}{
				"name":      "Putu Cempaka Giriyanti",
				"full_name": "Putu Cempaka Giriyanti",
				"nickname":  "Cempaka",
				"father":    "Bapak Giri",
				"mother":    "Ibu Yanti",
				"parents":   "Putri dari Bapak Giri & Ibu Yanti",
				"instagram": "cempakagiri",
				"photo_url": "https://storage.domain.com/photos/bride.jpg",
			}
			eventsData := map[string]interface{}{
				"holy_matrimony": map[string]interface{}{
					"title":      "Holy Matrimony",
					"date":       "2026-05-18",
					"time_range": "09:00 - 11:00 WITA",
					"location":   "Gereja Katedral Santo Roh Kudus",
					"address":    "Jl. Tukad Musi No. 1, Denpasar, Bali",
					"maps_url":   "https://maps.google.com/?q=...",
				},
				"reception": map[string]interface{}{
					"title":      "Wedding Reception",
					"date":       "2026-05-18",
					"time_range": "18:00 WITA - Selesai",
					"location":   "The Glass House Bali",
					"address":    "Jl. Danau Tamblingan No. 89, Sanur, Bali",
					"maps_url":   "https://maps.google.com/?q=...",
				},
			}
			var templateID *uuid.UUID
			if tmpl, ok := persistedTemplates["scarlet-sketch"]; ok {
				templateID = &tmpl.ID
			}

			themeData := map[string]interface{}{
				"template_id":        "scarlet-sketch",
				"template_component": "TemplateScarletSketch",
				"primary_color":      "#9E0F00",
			}
			storyData := "Perjalanan cinta kami dimulai di Denpasar hingga akhirnya memutuskan melangkah ke jenjang pernikahan."
			galleryData := []string{
				"https://storage.domain.com/gallery/photo1.jpg",
				"https://storage.domain.com/gallery/photo2.jpg",
				"https://storage.domain.com/gallery/photo3.jpg",
			}
			giftData := map[string]interface{}{
				"bank_name":      "BCA",
				"account_number": "127326134",
				"account_name":   "Putu Cempaka Giriyanti Aryawijaya",
				"qr_code_url":    nil,
			}
			giftsData := []map[string]interface{}{
				{
					"id":             "gift-1",
					"bank_name":      "BCA",
					"account_number": "127326134",
					"account_name":   "Putu Cempaka Giriyanti Aryawijaya",
					"account_holder": "Putu Cempaka Giriyanti Aryawijaya",
					"notes":          "Rekening Utama",
				},
			}

			inv := models.Invitation{
				OrderID:             po.ID,
				ClientID:            po.ClientID,
				PackageID:           po.PackageID,
				TemplateID:          templateID,
				Title:               "The Wedding of Mulyo & Cempaka",
				Slug:                &slugCandidate,
				Status:              "published",
				AudioURL:            &audioURL,
				VerseText:           &verseText,
				VerseRef:            &verseRef,
				CountdownTargetDate: &countdownDate,
				Events:              eventsData,
				Event:               eventsData,
				Groom:               groomData,
				Bride:               brideData,
				Theme:               themeData,
				Story:               storyData,
				Gallery:             galleryData,
				Gift:                giftData,
				Gifts:               giftsData,
			}

			var existingInv models.Invitation
			if err := tx.Where("order_id = ?", po.ID).First(&existingInv).Error; err == nil {
				existingInv.Title = inv.Title
				existingInv.Slug = inv.Slug
				existingInv.TemplateID = inv.TemplateID
				existingInv.Status = inv.Status
				existingInv.AudioURL = inv.AudioURL
				existingInv.VerseText = inv.VerseText
				existingInv.VerseRef = inv.VerseRef
				existingInv.CountdownTargetDate = inv.CountdownTargetDate
				existingInv.Events = inv.Events
				existingInv.Event = inv.Event
				existingInv.Groom = inv.Groom
				existingInv.Bride = inv.Bride
				existingInv.Theme = inv.Theme
				existingInv.Story = inv.Story
				existingInv.Gallery = inv.Gallery
				existingInv.Gift = inv.Gift
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
			if err := tx.Unscoped().Where("order_id = ?", po.ID).Delete(&models.Invitation{}).Error; err != nil {
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

		// Seed sample wishes for public guest book
		wishesData := []models.Wish{
			{
				InvitationID:     targetInvitationID,
				GuestName:        "Arthur",
				AttendanceStatus: "hadir",
				NumberOfGuests:   2,
				Message:          "Selamat menempuh hidup baru untuk Mulyo dan Cempaka! Semoga menjadi keluarga yang sakinah, mawaddah, warahmah.",
			},
			{
				InvitationID:     targetInvitationID,
				GuestName:        "Nanik",
				AttendanceStatus: "hadir",
				NumberOfGuests:   1,
				Message:          "Happy Wedding Mulyo & Cempaka! Bahagia dan langgeng selalu sampai kakek nenek yaa.",
			},
			{
				InvitationID:     targetInvitationID,
				GuestName:        "Budi Santoso",
				AttendanceStatus: "hadir",
				NumberOfGuests:   2,
				Message:          "Barakallahu lakuma wa baraka 'alaikuma wa jama'a bainakuma fii khoir. Selamat ya!",
			},
		}

		for _, w := range wishesData {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&w).Error; err != nil {
				return nil, fmt.Errorf("failed to seed wish: %w", err)
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
		"wishes_seeded":      3,
	}

	return result, nil
}
