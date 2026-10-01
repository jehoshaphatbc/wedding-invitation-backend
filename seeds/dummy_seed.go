package seeds

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

func SeedDummyData(db *gorm.DB) error {
	log.Println("==> Starting Hierarchical Database Seeder for Dummy Data...")

	// ---------------------------------------------------------
	// 1. SEED PACKAGES
	// ---------------------------------------------------------
	log.Println("--> 1. Seeding Packages (Silver, Gold, Platinum)...")

	packagesData := []struct {
		Name           string
		Price          float64
		FeaturesConfig models.FeaturesConfig
	}{
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
	for i, pd := range packagesData {
		var pkg models.Package
		err := db.Where("name = ?", pd.Name).First(&pkg).Error
		if err != nil {
			pkg = models.Package{
				Name:           pd.Name,
				Price:          pd.Price,
				FeaturesConfig: pd.FeaturesConfig,
			}
			if err := db.Create(&pkg).Error; err != nil {
				return fmt.Errorf("failed to create package %s: %w", pd.Name, err)
			}
			log.Printf("   Created package: %s (Rp %.0f)", pkg.Name, pkg.Price)
		} else {
			// Update features_config to match desired seed
			pkg.Price = pd.Price
			pkg.FeaturesConfig = pd.FeaturesConfig
			db.Save(&pkg)
			log.Printf("   Found existing package: %s", pkg.Name)
		}
		packages[i] = pkg
	}

	// ---------------------------------------------------------
	// 2. SEED CLIENTS (10 Realistic Indonesian Clients)
	// ---------------------------------------------------------
	log.Println("--> 2. Seeding 10 Indonesian Clients...")

	clientsData := []struct {
		Name     string
		Email    string
		Whatsapp string
	}{
		{"Raden Budi Santoso", "budi.santoso@gmail.com", "081298765432"},
		{"Siti Nurhaliza", "siti.nurhaliza@gmail.com", "085712345678"},
		{"Arya Pratama", "arya.pratama@gmail.com", "087811223344"},
		{"Dewi Lestari", "dewi.lestari@gmail.com", "081388776655"},
		{"Dimas Setiawan", "dimas.setiawan@gmail.com", "082199887766"},
		{"Anisa Rahmawati", "anisa.rahmawati@gmail.com", "085644332211"},
		{"Rizky Ramadhan", "rizky.ramadhan@gmail.com", "081233445566"},
		{"Putri Ayu Wulandari", "putri.wulandari@gmail.com", "087755667788"},
		{"Fajar Hidayat", "fajar.hidayat@gmail.com", "081922334455"},
		{"Mega Permata", "mega.permata@gmail.com", "085211223344"},
	}

	clients := make([]models.Client, len(clientsData))
	for i, cd := range clientsData {
		var client models.Client
		err := db.Where("email = ?", cd.Email).First(&client).Error
		if err != nil {
			client = models.Client{
				Name:     cd.Name,
				Email:    cd.Email,
				Whatsapp: cd.Whatsapp,
			}
			if err := db.Create(&client).Error; err != nil {
				return fmt.Errorf("failed to create client %s: %w", cd.Name, err)
			}
			log.Printf("   Created client: %s (%s)", client.Name, client.Email)
		} else {
			log.Printf("   Found existing client: %s", client.Name)
		}
		clients[i] = client
	}

	// ---------------------------------------------------------
	// 3. SEED ORDERS (15 Orders: 5 Unpaid, 8 Paid, 2 Expired)
	// ---------------------------------------------------------
	log.Println("--> 3. Seeding 15 Orders (5 Unpaid, 8 Paid, 2 Expired)...")

	orderPlan := []struct {
		InvoiceSuffix string
		ClientIdx     int
		PackageIdx    int
		Status        models.OrderStatus
	}{
		// 8 Paid Orders
		{"0001", 0, 1, models.OrderStatusPaid}, // Budi, Gold -> Paid (will have published invitation budi-ani)
		{"0003", 2, 2, models.OrderStatusPaid}, // Arya, Platinum -> Paid (will have published invitation arya-dewi)
		{"0004", 3, 1, models.OrderStatusPaid}, // Dewi, Gold -> Paid
		{"0006", 5, 1, models.OrderStatusPaid}, // Anisa, Gold -> Paid
		{"0007", 6, 2, models.OrderStatusPaid}, // Rizky, Platinum -> Paid
		{"0009", 8, 1, models.OrderStatusPaid}, // Fajar, Gold -> Paid
		{"0010", 9, 2, models.OrderStatusPaid}, // Mega, Platinum -> Paid
		{"0011", 0, 2, models.OrderStatusPaid}, // Budi (2nd order), Platinum -> Paid

		// 5 Unpaid Orders
		{"0002", 1, 0, models.OrderStatusUnpaid}, // Siti, Silver -> Unpaid
		{"0005", 4, 0, models.OrderStatusUnpaid}, // Dimas, Silver -> Unpaid
		{"0012", 1, 1, models.OrderStatusUnpaid}, // Siti, Gold -> Unpaid
		{"0013", 2, 0, models.OrderStatusUnpaid}, // Arya, Silver -> Unpaid
		{"0014", 3, 2, models.OrderStatusUnpaid}, // Dewi, Platinum -> Unpaid

		// 2 Expired Orders
		{"0008", 7, 0, models.OrderStatusExpired}, // Putri, Silver -> Expired
		{"0015", 4, 1, models.OrderStatusExpired}, // Dimas, Gold -> Expired
	}

	paidOrders := make([]models.Order, 0)

	for _, op := range orderPlan {
		client := clients[op.ClientIdx]
		pkg := packages[op.PackageIdx]
		invNum := fmt.Sprintf("INV-20261001-%s", op.InvoiceSuffix)

		var order models.Order
		err := db.Where("invoice_number = ?", invNum).First(&order).Error
		if err != nil {
			var formToken, scannerToken string
			if op.Status == models.OrderStatusPaid {
				formToken = "form_" + uuid.New().String()
				scannerToken = "scan_" + uuid.New().String()
			}

			order = models.Order{
				InvoiceNumber: invNum,
				ClientID:      client.ID,
				PackageID:     pkg.ID,
				TotalAmount:   pkg.Price,
				Status:        op.Status,
				PaymentURL:    fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", uuid.New().String()),
				FormToken:     formToken,
				ScannerToken:  scannerToken,
			}

			if err := db.Create(&order).Error; err != nil {
				return fmt.Errorf("failed to create order %s: %w", invNum, err)
			}
			log.Printf("   Created order: %s [%s] Client: %s, Pkg: %s", order.InvoiceNumber, order.Status, client.Name, pkg.Name)
		} else {
			log.Printf("   Found existing order: %s [%s]", order.InvoiceNumber, order.Status)
		}

		if op.Status == models.OrderStatusPaid {
			paidOrders = append(paidOrders, order)
		}
	}

	// ---------------------------------------------------------
	// 4. SEED INVITATIONS (8 Invitations for Paid Orders)
	// ---------------------------------------------------------
	log.Println("--> 4. Seeding Invitations for Paid Orders (2 Published, 6 Draft)...")

	var publishedInvitation *models.Invitation

	for i, order := range paidOrders {
		var invitation models.Invitation
		err := db.Where("order_id = ?", order.ID).First(&invitation).Error
		if err != nil {
			var status string
			var slug *string
			var title string

			if i == 0 {
				// Published Invitation 1
				status = "published"
				s := "budi-ani"
				slug = &s
				title = "The Wedding of Budi & Ani"
			} else if i == 1 {
				// Published Invitation 2
				status = "published"
				s := "arya-dewi"
				slug = &s
				title = "The Wedding of Arya & Dewi"
			} else {
				// Draft Invitations
				status = "draft"
				title = fmt.Sprintf("Draft Undangan Pernikahan #%d", i+1)
			}

			invitation = models.Invitation{
				OrderID:   order.ID,
				ClientID:  order.ClientID,
				PackageID: order.PackageID,
				Title:     title,
				Slug:      slug,
				Status:    status,
			}

			if err := db.Create(&invitation).Error; err != nil {
				return fmt.Errorf("failed to create invitation for order %s: %w", order.InvoiceNumber, err)
			}
			log.Printf("   Created invitation: %s (Status: %s, Slug: %v)", invitation.Title, invitation.Status, slug)
		} else {
			log.Printf("   Found existing invitation: %s", invitation.Title)
		}

		if invitation.Slug != nil && *invitation.Slug == "budi-ani" {
			publishedInvitation = &invitation
		}
	}

	// Fallback to first invitation if budi-ani pointer wasn't set
	if publishedInvitation == nil && len(paidOrders) > 0 {
		var inv models.Invitation
		if err := db.Where("status = ?", "published").First(&inv).Error; err == nil {
			publishedInvitation = &inv
		}
	}

	// ---------------------------------------------------------
	// 5. SEED GUESTS (30 Guests for 1 Published Invitation)
	// ---------------------------------------------------------
	if publishedInvitation != nil {
		log.Printf("--> 5. Seeding 30 Guests for Invitation ID: %s (Slug: %v)...", publishedInvitation.ID, publishedInvitation.Slug)

		guestSeedData := []struct {
			Name             string
			Phone            string
			RSVPStatus       string
			ActualAttendance bool
		}{
			// 10 'hadir' (5 actual_attendance: true, 5 actual_attendance: false)
			{"Hendra Gunawan", "081211112222", "hadir", true},
			{"Eko Prasetyo", "081222223333", "hadir", true},
			{"Bambang Suryono", "081233334444", "hadir", true},
			{"Agus Wijaya", "081244445555", "hadir", true},
			{"Dian Sastrowardoyo", "081255556666", "hadir", true},
			{"Tri Haryanto", "081266667777", "hadir", false},
			{"Wahyu Wibowo", "081277778888", "hadir", false},
			{"Sri Wahyuni", "081288889999", "hadir", false},
			{"Indah Permatasari", "081299990000", "hadir", false},
			{"Bayu Nugroho", "081311112222", "hadir", false},

			// 5 'tidak_hadir' (all actual_attendance: false)
			{"Rina Nose", "081322223333", "tidak_hadir", false},
			{"Doni Kusuma", "081333334444", "tidak_hadir", false},
			{"Maya Safira", "081344445555", "tidak_hadir", false},
			{"Gilang Ramadhan", "081355556666", "tidak_hadir", false},
			{"Tari Melinda", "081366667777", "tidak_hadir", false},

			// 15 'pending' (all actual_attendance: false)
			{"Aditya Pratama", "081377778888", "pending", false},
			{"Bagus Triadi", "081388889999", "pending", false},
			{"Citra Kirana", "081399990000", "pending", false},
			{"Danang Sutrisno", "081411112222", "pending", false},
			{"Erwin Saputra", "081422223333", "pending", false},
			{"Farhan Alamsyah", "081433334444", "pending", false},
			{"Gita Gutawa", "081444445555", "pending", false},
			{"Hesti Purwadinata", "081455556666", "pending", false},
			{"Irfan Hakim", "081466667777", "pending", false},
			{"Joko Widodo", "081477778888", "pending", false},
			{"Kartika Sari", "081488889999", "pending", false},
			{"Lukman Sardi", "081499990000", "pending", false},
			{"Maulana Malik", "081511112222", "pending", false},
			{"Nina Zatulini", "081522223333", "pending", false},
			{"Oscar Lawalata", "081533334444", "pending", false},
		}

		now := time.Now().Add(-2 * time.Hour) // Attended 2 hours ago

		for i, gd := range guestSeedData {
			qrToken := fmt.Sprintf("QR-%s-%02d", publishedInvitation.ID.String()[:8], i+1)

			var guest models.Guest
			err := db.Where("qr_token = ?", qrToken).First(&guest).Error
			if err != nil {
				phone := gd.Phone
				var attTime *time.Time
				if gd.ActualAttendance {
					tTime := now.Add(time.Duration(i*5) * time.Minute)
					attTime = &tTime
				}

				guest = models.Guest{
					InvitationID:     publishedInvitation.ID,
					Name:             gd.Name,
					Phone:            &phone,
					Pax:              1,
					QRToken:          qrToken,
					RSVPStatus:       gd.RSVPStatus,
					ActualAttendance: gd.ActualAttendance,
					AttendanceTime:   attTime,
				}

				if err := db.Create(&guest).Error; err != nil {
					return fmt.Errorf("failed to create guest %s: %w", gd.Name, err)
				}
				log.Printf("   Created guest #%02d: %s [RSVP: %s, Attended: %v, QR: %s]", i+1, guest.Name, guest.RSVPStatus, guest.ActualAttendance, guest.QRToken)
			} else {
				log.Printf("   Found existing guest: %s", guest.Name)
			}
		}
	}

	log.Println("==> Database Seeder for Dummy Data Completed Successfully!")
	return nil
}
