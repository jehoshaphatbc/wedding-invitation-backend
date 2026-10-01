package services

import (
	"errors"

	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type ClientPortalService struct {
	orderRepo      repositories.OrderRepository
	invitationRepo repositories.InvitationRepository
	auditRepo      repositories.AuditLogRepository
}

func NewClientPortalService(
	orderRepo repositories.OrderRepository,
	invitationRepo repositories.InvitationRepository,
	auditRepo repositories.AuditLogRepository,
) *ClientPortalService {
	return &ClientPortalService{
		orderRepo:      orderRepo,
		invitationRepo: invitationRepo,
		auditRepo:      auditRepo,
	}
}

// VerifyToken verifies the magic link form_token, loads order, client, package features_config, and invitation.
func (s *ClientPortalService) VerifyToken(token string) (*models.ClientAuthVerifyResponse, error) {
	if token == "" {
		return nil, errors.New("form token is required")
	}

	order, err := s.orderRepo.FindByFormToken(token)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("unauthorized: invalid or expired magic link token")
	}

	invitation, err := s.invitationRepo.FindByOrderID(order.ID)
	if err != nil {
		return nil, err
	}

	// Auto-create draft invitation if not exists yet
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

	res := &models.ClientAuthVerifyResponse{
		Order: models.ClientOrderSummary{
			ID:            order.ID,
			InvoiceNumber: order.InvoiceNumber,
			Status:        order.Status,
			TotalAmount:   order.TotalAmount,
			FormToken:     order.FormToken,
			ScannerToken:  order.ScannerToken,
		},
		Invitation: invitation,
	}

	if order.Client != nil {
		res.Client = &models.ClientSummary{
			ID:       order.Client.ID,
			Name:     order.Client.Name,
			Email:    order.Client.Email,
			Whatsapp: order.Client.Whatsapp,
		}
	}

	if order.Package != nil {
		res.Package = &models.ClientPackageSummary{
			ID:             order.Package.ID,
			Name:           order.Package.Name,
			FeaturesConfig: order.Package.FeaturesConfig,
		}
	}

	return res, nil
}

// UpdateInvitation updates invitation content fields for a given order.
func (s *ClientPortalService) UpdateInvitation(orderID uuid.UUID, req models.UpdateClientInvitationRequest, ip, userAgent string) (*models.Invitation, error) {
	invitation, err := s.invitationRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	if invitation == nil {
		invitation = &models.Invitation{
			OrderID: orderID,
			Title:   "Draft Undangan",
			Status:  "draft",
		}
		if req.Title != nil {
			invitation.Title = *req.Title
		}
		if err := s.invitationRepo.Create(invitation); err != nil {
			return nil, err
		}
	}

	if req.Title != nil {
		invitation.Title = *req.Title
	}
	if req.Slug != nil {
		invitation.Slug = req.Slug
	}
	if req.GroomData != nil {
		invitation.GroomData = req.GroomData
	}
	if req.BrideData != nil {
		invitation.BrideData = req.BrideData
	}
	if req.EventsData != nil {
		invitation.EventsData = req.EventsData
	}
	if req.StoryData != nil {
		invitation.StoryData = req.StoryData
	}
	if req.GalleryURLs != nil {
		invitation.GalleryURLs = req.GalleryURLs
	}

	if err := s.invitationRepo.Update(invitation); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		s.auditRepo.Create(&models.AuditLog{
			Action:       "client_invitation.updated",
			ResourceType: "invitations",
			ResourceID:   &invitation.ID,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
		})
	}

	return invitation, nil
}
