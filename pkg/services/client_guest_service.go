package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type ClientGuestService struct {
	guestRepo      repositories.GuestRepository
	invitationRepo repositories.InvitationRepository
	auditRepo      repositories.AuditLogRepository
}

func NewClientGuestService(
	guestRepo repositories.GuestRepository,
	invitationRepo repositories.InvitationRepository,
	auditRepo repositories.AuditLogRepository,
) *ClientGuestService {
	return &ClientGuestService{
		guestRepo:      guestRepo,
		invitationRepo: invitationRepo,
		auditRepo:      auditRepo,
	}
}

// getOrCreateInvitation finds or auto-creates an invitation for the given order
func (s *ClientGuestService) getOrCreateInvitation(order *models.Order) (*models.Invitation, error) {
	if order == nil {
		return nil, errors.New("order is required")
	}

	invitation, err := s.invitationRepo.FindByOrderID(order.ID)
	if err != nil {
		return nil, err
	}

	if invitation == nil {
		title := "Draft Undangan"
		if order.Client != nil && order.Client.Name != "" {
			title = "Undangan " + order.Client.Name
		}
		invitation = &models.Invitation{
			OrderID:   order.ID,
			ClientID:  order.ClientID,
			PackageID: order.PackageID,
			Title:     title,
			Status:    "draft",
		}
		if err := s.invitationRepo.Create(invitation); err != nil {
			return nil, err
		}
	}

	return invitation, nil
}

// GetGuests returns all guests for the order's invitation along with RSVP aggregations
func (s *ClientGuestService) GetGuests(order *models.Order) (*models.ClientGuestListResponse, error) {
	invitation, err := s.getOrCreateInvitation(order)
	if err != nil {
		return nil, err
	}

	guests, err := s.guestRepo.FindByInvitationID(invitation.ID)
	if err != nil {
		return nil, err
	}

	var totalHadir int64
	var totalTidakHadir int64
	var totalPending int64

	for _, g := range guests {
		switch strings.ToLower(strings.TrimSpace(g.RSVPStatus)) {
		case "hadir":
			totalHadir++
		case "tidak_hadir":
			totalTidakHadir++
		default:
			totalPending++
		}
	}

	totalGuests := int64(len(guests))

	res := &models.ClientGuestListResponse{
		Guests:          guests,
		TotalGuests:     totalGuests,
		TotalHadir:      totalHadir,
		TotalTidakHadir: totalTidakHadir,
		TotalPending:    totalPending,
		Summary: models.GuestSummaryResponse{
			TotalGuests:     totalGuests,
			TotalHadir:      totalHadir,
			TotalTidakHadir: totalTidakHadir,
			TotalPending:    totalPending,
		},
	}

	return res, nil
}

// BulkCreateGuests adds multiple guests for the client's invitation
func (s *ClientGuestService) BulkCreateGuests(order *models.Order, req models.BulkCreateGuestsRequest, ip, userAgent string) ([]models.Guest, error) {
	invitation, err := s.getOrCreateInvitation(order)
	if err != nil {
		return nil, err
	}

	var guestsToCreate []models.Guest
	for _, rawName := range req.Names {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}

		qrToken := "qr_" + uuid.New().String()
		guest := models.Guest{
			ID:           uuid.New(),
			InvitationID: invitation.ID,
			Name:         name,
			Pax:          1,
			QRToken:      qrToken,
			RSVPStatus:   "pending",
		}
		guestsToCreate = append(guestsToCreate, guest)
	}

	if len(guestsToCreate) == 0 {
		return nil, errors.New("names array must contain at least one non-empty name")
	}

	if err := s.guestRepo.BulkCreate(guestsToCreate); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		s.auditRepo.Create(&models.AuditLog{
			Action:       "client_guests.bulk_created",
			ResourceType: "guests",
			ResourceID:   &invitation.ID,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
		})
	}

	return guestsToCreate, nil
}

// UpdateGuest updates an existing guest belonging to the client's invitation
func (s *ClientGuestService) UpdateGuest(order *models.Order, guestID uuid.UUID, req models.UpdateGuestRequest, ip, userAgent string) (*models.Guest, error) {
	invitation, err := s.getOrCreateInvitation(order)
	if err != nil {
		return nil, err
	}

	guest, err := s.guestRepo.FindByIDAndInvitationID(guestID, invitation.ID)
	if err != nil {
		return nil, err
	}
	if guest == nil {
		return nil, errors.New("tamu tidak ditemukan atau Anda tidak memiliki akses ke tamu ini")
	}

	// Update name if provided (support name or guest_name)
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		guest.Name = strings.TrimSpace(*req.Name)
	} else if req.GuestName != nil && strings.TrimSpace(*req.GuestName) != "" {
		guest.Name = strings.TrimSpace(*req.GuestName)
	}

	if req.Phone != nil {
		guest.Phone = req.Phone
	}
	if req.Pax != nil && *req.Pax > 0 {
		guest.Pax = *req.Pax
	}
	if req.RSVPStatus != nil && strings.TrimSpace(*req.RSVPStatus) != "" {
		guest.RSVPStatus = strings.TrimSpace(*req.RSVPStatus)
	}

	if err := s.guestRepo.Update(guest); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		s.auditRepo.Create(&models.AuditLog{
			Action:       "client_guest.updated",
			ResourceType: "guests",
			ResourceID:   &guest.ID,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
		})
	}

	return guest, nil
}

// DeleteGuest removes a guest belonging to the client's invitation
func (s *ClientGuestService) DeleteGuest(order *models.Order, guestID uuid.UUID, ip, userAgent string) error {
	invitation, err := s.getOrCreateInvitation(order)
	if err != nil {
		return err
	}

	guest, err := s.guestRepo.FindByIDAndInvitationID(guestID, invitation.ID)
	if err != nil {
		return err
	}
	if guest == nil {
		return errors.New("tamu tidak ditemukan atau Anda tidak memiliki akses ke tamu ini")
	}

	if err := s.guestRepo.Delete(guestID, invitation.ID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		s.auditRepo.Create(&models.AuditLog{
			Action:       "client_guest.deleted",
			ResourceType: "guests",
			ResourceID:   &guestID,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
		})
	}

	return nil
}
