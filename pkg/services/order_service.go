package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type OrderService struct {
	orderRepo      repositories.OrderRepository
	clientRepo     repositories.ClientRepository
	packageRepo    repositories.PackageRepository
	invitationRepo repositories.InvitationRepository
	auditRepo      repositories.AuditLogRepository
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	clientRepo repositories.ClientRepository,
	packageRepo repositories.PackageRepository,
	invitationRepo repositories.InvitationRepository,
	auditRepo repositories.AuditLogRepository,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		clientRepo:     clientRepo,
		packageRepo:    packageRepo,
		invitationRepo: invitationRepo,
		auditRepo:      auditRepo,
	}
}

func (s *OrderService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
	if s.auditRepo == nil {
		return
	}
	s.auditRepo.Create(&models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
	})
}

func generateRandomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return uuid.New().String()[:n*2]
	}
	return hex.EncodeToString(bytes)
}

func generateInvoiceNumber() string {
	dateStr := time.Now().Format("20060102")
	randomPart := strings.ToUpper(generateRandomHex(3)) // 6 chars hex
	return fmt.Sprintf("INV-%s-%s", dateStr, randomPart)
}

// Public Checkout: cari/buat client, buat order (unpaid), kembalikan payment_url (mockup Midtrans)
func (s *OrderService) Checkout(req models.PublicCheckoutRequest, ip, userAgent string) (*models.Order, *models.Client, *models.Package, error) {
	// 1. Verify package exists
	pkg, err := s.packageRepo.FindByID(req.PackageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, errors.New("package not found")
		}
		return nil, nil, nil, err
	}

	// 2. Find or create client
	client, err := s.clientRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, nil, nil, err
	}

	if client == nil {
		client = &models.Client{
			Name:     req.Name,
			Email:    req.Email,
			Whatsapp: req.Whatsapp,
		}
		if err := s.clientRepo.Create(client); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create client: %w", err)
		}
		s.auditLog(nil, "client.created_via_checkout", "clients", &client.ID, ip, userAgent)
	} else {
		// Update name/whatsapp if changed
		client.Name = req.Name
		client.Whatsapp = req.Whatsapp
		_ = s.clientRepo.Update(client)
	}

	// 3. Generate unique invoice number and mock Midtrans payment URL
	invoiceNumber := generateInvoiceNumber()
	mockSnapToken := uuid.New().String()
	paymentURL := fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", mockSnapToken)

	order := &models.Order{
		InvoiceNumber: invoiceNumber,
		ClientID:      client.ID,
		PackageID:     pkg.ID,
		TotalAmount:   pkg.Price,
		Status:        models.OrderStatusUnpaid,
		PaymentURL:    paymentURL,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create order: %w", err)
	}

	order.Client = client
	order.Package = pkg

	s.auditLog(nil, "order.checkout_created", "orders", &order.ID, ip, userAgent)
	return order, client, pkg, nil
}

// Public Webhook: update order status, generate form_token & scanner_token, otomatis buat 1 row data draft di invitations
func (s *OrderService) HandlePaymentWebhook(req models.PaymentWebhookRequest, ip, userAgent string) (*models.Order, *models.Invitation, error) {
	if req.OrderID == "" {
		return nil, nil, errors.New("order_id is required")
	}

	// 1. Locate order by invoice_number or UUID
	order, err := s.orderRepo.FindByInvoiceNumber(req.OrderID)
	if err != nil {
		return nil, nil, err
	}

	if order == nil {
		if orderUUID, parseErr := uuid.Parse(req.OrderID); parseErr == nil {
			order, err = s.orderRepo.FindByID(orderUUID)
			if err != nil {
				return nil, nil, errors.New("order not found")
			}
		}
	}

	if order == nil {
		return nil, nil, errors.New("order not found")
	}

	// 2. Determine target status from webhook notification
	targetStatus := models.OrderStatusUnpaid
	ts := strings.ToLower(req.TransactionStatus)
	st := strings.ToLower(req.Status)

	if ts == "settlement" || ts == "capture" || st == "paid" {
		targetStatus = models.OrderStatusPaid
	} else if ts == "expire" || ts == "cancel" || ts == "deny" || st == "expired" {
		targetStatus = models.OrderStatusExpired
	}

	var invitation *models.Invitation

	// 3. Process status update
	if targetStatus == models.OrderStatusPaid {
		// Idempotency check: if already paid, ensure tokens and invitation exist and return
		if order.Status == models.OrderStatusPaid {
			invitation, _ = s.invitationRepo.FindByOrderID(order.ID)
			return order, invitation, nil
		}

		order.Status = models.OrderStatusPaid
		if order.FormToken == "" {
			order.FormToken = "form_" + generateRandomHex(16)
		}
		if order.ScannerToken == "" {
			order.ScannerToken = "scan_" + generateRandomHex(16)
		}

		if err := s.orderRepo.Update(order); err != nil {
			return nil, nil, fmt.Errorf("failed to update order: %w", err)
		}

		// Otomatis buat 1 row draft di tabel invitations jika belum ada
		invitation, err = s.invitationRepo.FindByOrderID(order.ID)
		if err != nil {
			return nil, nil, err
		}

		if invitation == nil {
			invitation = &models.Invitation{
				OrderID:   order.ID,
				ClientID:  order.ClientID,
				PackageID: order.PackageID,
				Title:     "Draft Undangan",
				Status:    "draft",
			}
			if err := s.invitationRepo.Create(invitation); err != nil {
				return nil, nil, fmt.Errorf("failed to create draft invitation: %w", err)
			}
			s.auditLog(nil, "invitation.auto_created", "invitations", &invitation.ID, ip, userAgent)
		}

		s.auditLog(nil, "order.paid", "orders", &order.ID, ip, userAgent)
	} else if targetStatus == models.OrderStatusExpired {
		order.Status = models.OrderStatusExpired
		if err := s.orderRepo.Update(order); err != nil {
			return nil, nil, fmt.Errorf("failed to update order: %w", err)
		}
		s.auditLog(nil, "order.expired", "orders", &order.ID, ip, userAgent)
	}

	return order, invitation, nil
}

// Admin Standard Order CRUD
func (s *OrderService) GetAllOrders(page, perPage int, search, status string, isTrashed bool) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 && perPage != -1 {
		perPage = 20
	}
	return s.orderRepo.FindAll(page, perPage, search, status, isTrashed)
}

func (s *OrderService) GetOrderByID(id uuid.UUID) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return order, nil
}

func (s *OrderService) UpdateOrder(id uuid.UUID, req models.UpdateOrderRequest, ip, userAgent string) (*models.Order, error) {
	order, err := s.GetOrderByID(id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		order.Status = *req.Status
		if *req.Status == models.OrderStatusPaid {
			if order.FormToken == "" {
				order.FormToken = "form_" + generateRandomHex(16)
			}
			if order.ScannerToken == "" {
				order.ScannerToken = "scan_" + generateRandomHex(16)
			}

			// Ensure invitation is created if paid
			invitation, _ := s.invitationRepo.FindByOrderID(order.ID)
			if invitation == nil {
				inv := &models.Invitation{
					OrderID:   order.ID,
					ClientID:  order.ClientID,
					PackageID: order.PackageID,
					Title:     "Draft Undangan",
					Status:    "draft",
				}
				_ = s.invitationRepo.Create(inv)
			}
		}
	}

	if req.TotalAmount != nil {
		order.TotalAmount = *req.TotalAmount
	}

	if err := s.orderRepo.Update(order); err != nil {
		return nil, err
	}

	s.auditLog(nil, "order.updated", "orders", &order.ID, ip, userAgent)
	return order, nil
}

func (s *OrderService) DeleteOrder(id uuid.UUID, ip, userAgent string) error {
	order, err := s.GetOrderByID(id)
	if err != nil {
		return err
	}

	if err := s.orderRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "order.deleted", "orders", &order.ID, ip, userAgent)
	return nil
}

func (s *OrderService) RestoreOrder(id uuid.UUID, ip, userAgent string) error {
	if err := s.orderRepo.Restore(id); err != nil {
		return err
	}

	s.auditLog(nil, "order.restored", "orders", &id, ip, userAgent)
	return nil
}

func (s *OrderService) ForceDeleteOrder(id uuid.UUID, ip, userAgent string) error {
	if err := s.orderRepo.ForceDelete(id); err != nil {
		return err
	}

	s.auditLog(nil, "order.force_deleted", "orders", &id, ip, userAgent)
	return nil
}

func (s *OrderService) GetTrashedOrders(page, perPage int, search, sort, order string) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return s.orderRepo.FindTrashedAll(page, perPage, search, sort, order)
}
