package services

import (
	"errors"
	"fmt"
	"strings"

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

// extractTemplateID extracts template_id from theme map/object
func extractTemplateID(theme interface{}) string {
	if theme == nil {
		return ""
	}
	if m, ok := theme.(map[string]interface{}); ok {
		if tid, exists := m["template_id"]; exists && tid != nil {
			return strings.TrimSpace(fmt.Sprintf("%v", tid))
		}
	}
	return ""
}

// VerifyToken verifies the magic link form_token, loads order, client, package features_config,
// and determines if setup is completed. Returns invitation=null if client hasn't completed setup.
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

	isSetupCompleted := false
	if invitation != nil {
		hasGroom := invitation.Groom != nil || invitation.GroomData != nil
		hasBride := invitation.Bride != nil || invitation.BrideData != nil
		hasEvent := invitation.Event != nil || invitation.EventsData != nil

		if invitation.Status == "published" || (hasGroom && hasBride) || hasEvent {
			isSetupCompleted = true
		}
	}

	var returnedInvitation *models.Invitation
	if isSetupCompleted {
		returnedInvitation = invitation
	}

	res := &models.ClientAuthVerifyResponse{
		Valid:            true,
		Token:            token,
		IsSetupCompleted: isSetupCompleted,
		Order: models.ClientOrderSummary{
			ID:            order.ID,
			InvoiceNumber: order.InvoiceNumber,
			Status:        order.Status,
			TotalAmount:   order.TotalAmount,
			FormToken:     order.FormToken,
			ScannerToken:  order.ScannerToken,
		},
		Invitation: returnedInvitation,
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

// UpdateInvitation performs an upsert of invitation data for a given order,
// enforces theme locking, and sets status to published upon form completion.
func (s *ClientPortalService) UpdateInvitation(orderID, clientID, packageID uuid.UUID, req models.UpdateClientInvitationRequest, ip, userAgent string) (*models.Invitation, error) {
	invitation, err := s.invitationRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	// 1. Theme Immutability Enforcement
	existingTemplateID := ""
	if invitation != nil && invitation.Theme != nil {
		existingTemplateID = extractTemplateID(invitation.Theme)
	}

	if existingTemplateID != "" && req.Theme != nil {
		newTemplateID := extractTemplateID(req.Theme)
		if newTemplateID != "" && newTemplateID != existingTemplateID {
			return nil, errors.New("Template theme is permanently locked and cannot be changed after initial setup.")
		}
		// If payload didn't specify template_id, preserve existing template_id
		if m, ok := req.Theme.(map[string]interface{}); ok {
			if _, hasTID := m["template_id"]; !hasTID || m["template_id"] == nil {
				m["template_id"] = existingTemplateID
			}
		}
	}

	isNew := false
	if invitation == nil {
		isNew = true
		title := "Draft Undangan"
		if req.Title != nil && *req.Title != "" {
			title = *req.Title
		}
		invitation = &models.Invitation{
			OrderID:   orderID,
			ClientID:  clientID,
			PackageID: packageID,
			Title:     title,
			Status:    "draft",
		}
	}

	if req.Title != nil && *req.Title != "" {
		invitation.Title = *req.Title
	}
	if req.Slug != nil && *req.Slug != "" {
		invitation.Slug = req.Slug
	}

	// Groom
	if req.Groom != nil {
		invitation.Groom = req.Groom
		if m, ok := req.Groom.(map[string]interface{}); ok {
			invitation.GroomData = m
		}
	} else if req.GroomData != nil {
		invitation.Groom = req.GroomData
		invitation.GroomData = req.GroomData
	}

	// Bride
	if req.Bride != nil {
		invitation.Bride = req.Bride
		if m, ok := req.Bride.(map[string]interface{}); ok {
			invitation.BrideData = m
		}
	} else if req.BrideData != nil {
		invitation.Bride = req.BrideData
		invitation.BrideData = req.BrideData
	}

	// Event
	if req.Event != nil {
		invitation.Event = req.Event
		invitation.EventsData = req.Event
	} else if req.EventsData != nil {
		invitation.Event = req.EventsData
		invitation.EventsData = req.EventsData
	}

	// Theme
	if req.Theme != nil {
		invitation.Theme = req.Theme
	}

	// Story
	if req.Story != nil {
		invitation.Story = req.Story
		invitation.StoryData = req.Story
	} else if req.StoryData != nil {
		invitation.Story = req.StoryData
		invitation.StoryData = req.StoryData
	}

	// Gallery
	if req.Gallery != nil {
		invitation.Gallery = req.Gallery
		if urls, ok := req.Gallery.([]string); ok {
			invitation.GalleryURLs = urls
		} else if slice, ok := req.Gallery.([]interface{}); ok {
			var strList []string
			for _, item := range slice {
				if s, ok := item.(string); ok {
					strList = append(strList, s)
				}
			}
			invitation.GalleryURLs = strList
		}
	} else if req.GalleryURLs != nil {
		invitation.Gallery = req.GalleryURLs
		invitation.GalleryURLs = req.GalleryURLs
	}

	// Gift & Gifts
	if req.Gift != nil {
		invitation.Gift = req.Gift
	}
	if req.Gifts != nil {
		invitation.Gifts = req.Gifts
	}

	// 2. Lifecycle Status: transition to "published" once client submits groom/bride/event
	hasGroom := invitation.Groom != nil || invitation.GroomData != nil
	hasBride := invitation.Bride != nil || invitation.BrideData != nil
	hasEvent := invitation.Event != nil || invitation.EventsData != nil

	if hasGroom || hasBride || hasEvent {
		invitation.Status = "published"
	}

	if isNew {
		if err := s.invitationRepo.Create(invitation); err != nil {
			return nil, err
		}
	} else {
		if err := s.invitationRepo.Update(invitation); err != nil {
			return nil, err
		}
	}

	if s.auditRepo != nil {
		action := "client_invitation.updated"
		if isNew {
			action = "client_invitation.created"
		}
		s.auditRepo.Create(&models.AuditLog{
			Action:       action,
			ResourceType: "invitations",
			ResourceID:   &invitation.ID,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
		})
	}

	return invitation, nil
}
