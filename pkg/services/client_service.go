package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type ClientService struct {
	clientRepo repositories.ClientRepository
	auditRepo  repositories.AuditLogRepository
}

func NewClientService(clientRepo repositories.ClientRepository, auditRepo repositories.AuditLogRepository) *ClientService {
	return &ClientService{
		clientRepo: clientRepo,
		auditRepo:  auditRepo,
	}
}

func (s *ClientService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
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

func (s *ClientService) CreateClient(req models.CreateClientRequest, ip, userAgent string) (*models.Client, error) {
	existing, err := s.clientRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("client with this email already exists")
	}

	client := &models.Client{
		Name:     req.Name,
		Email:    req.Email,
		Whatsapp: req.Whatsapp,
	}

	if err := s.clientRepo.Create(client); err != nil {
		return nil, err
	}

	s.auditLog(nil, "client.created", "clients", &client.ID, ip, userAgent)
	return client, nil
}

func (s *ClientService) GetAllClients(page, perPage int, search string, isTrashed bool) ([]models.Client, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 && perPage != -1 {
		perPage = 20
	}
	return s.clientRepo.FindAll(page, perPage, search, isTrashed)
}

func (s *ClientService) GetClientByID(id uuid.UUID) (*models.Client, error) {
	client, err := s.clientRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	return client, nil
}

func (s *ClientService) UpdateClient(id uuid.UUID, req models.UpdateClientRequest, ip, userAgent string) (*models.Client, error) {
	client, err := s.GetClientByID(id)
	if err != nil {
		return nil, err
	}

	if req.Email != nil && *req.Email != client.Email {
		existing, err := s.clientRepo.FindByEmail(*req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != client.ID {
			return nil, errors.New("client with this email already exists")
		}
		client.Email = *req.Email
	}

	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.Whatsapp != nil {
		client.Whatsapp = *req.Whatsapp
	}

	if err := s.clientRepo.Update(client); err != nil {
		return nil, err
	}

	s.auditLog(nil, "client.updated", "clients", &client.ID, ip, userAgent)
	return client, nil
}

func (s *ClientService) DeleteClient(id uuid.UUID, ip, userAgent string) error {
	client, err := s.GetClientByID(id)
	if err != nil {
		return err
	}

	if err := s.clientRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "client.deleted", "clients", &client.ID, ip, userAgent)
	return nil
}

func (s *ClientService) RestoreClient(id uuid.UUID, ip, userAgent string) error {
	if err := s.clientRepo.Restore(id); err != nil {
		return err
	}

	s.auditLog(nil, "client.restored", "clients", &id, ip, userAgent)
	return nil
}

func (s *ClientService) ForceDeleteClient(id uuid.UUID, ip, userAgent string) error {
	if err := s.clientRepo.ForceDelete(id); err != nil {
		return err
	}

	s.auditLog(nil, "client.force_deleted", "clients", &id, ip, userAgent)
	return nil
}

func (s *ClientService) GetTrashedClients(page, perPage int, search, sort, order string) ([]models.Client, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return s.clientRepo.FindTrashedAll(page, perPage, search, sort, order)
}
