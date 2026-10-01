package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type TemplateRequest struct {
	Name          string  `json:"name" form:"name" binding:"required"`
	NuxtComponent string  `json:"nuxt_component" form:"nuxt_component" binding:"required"`
	ThumbnailURL  *string `json:"thumbnail_url" form:"thumbnail_url"`
	IsActive      *bool   `json:"is_active" form:"is_active"`
}

type TemplateService struct {
	templateRepo repositories.TemplateRepository
	auditRepo    repositories.AuditLogRepository
}

func NewTemplateService(templateRepo repositories.TemplateRepository, auditRepo repositories.AuditLogRepository) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		auditRepo:    auditRepo,
	}
}

func (s *TemplateService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
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

func (s *TemplateService) CreateTemplate(req TemplateRequest) (*models.Template, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	template := &models.Template{
		Name:          req.Name,
		NuxtComponent: req.NuxtComponent,
		ThumbnailURL:  req.ThumbnailURL,
		IsActive:      isActive,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *TemplateService) GetAllTemplates(page, perPage int, search string, isActive *bool, isTrashed bool) ([]models.Template, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 && perPage != -1 {
		perPage = 20
	}
	return s.templateRepo.FindAll(page, perPage, search, isActive, isTrashed)
}

func (s *TemplateService) GetTemplateByID(id uuid.UUID) (*models.Template, error) {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return template, nil
}

func (s *TemplateService) UpdateTemplate(id uuid.UUID, req TemplateRequest) (*models.Template, error) {
	template, err := s.GetTemplateByID(id)
	if err != nil {
		return nil, err
	}

	template.Name = req.Name
	template.NuxtComponent = req.NuxtComponent
	if req.ThumbnailURL != nil {
		if *req.ThumbnailURL == "" {
			template.ThumbnailURL = nil
		} else {
			template.ThumbnailURL = req.ThumbnailURL
		}
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}

	if err := s.templateRepo.Update(template); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *TemplateService) DeleteTemplate(id uuid.UUID, ip, userAgent string) error {
	template, err := s.GetTemplateByID(id)
	if err != nil {
		return err
	}

	if err := s.templateRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "template.deleted", "templates", &template.ID, ip, userAgent)
	return nil
}

func (s *TemplateService) GetTrashedTemplates(page, perPage int, search, sort, order string) ([]models.Template, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.templateRepo.FindTrashedAll(page, perPage, search, sort, order)
}

func (s *TemplateService) RestoreTemplate(id uuid.UUID, ip, userAgent string) error {
	err := s.templateRepo.Restore(id)
	if err == nil {
		s.auditLog(nil, "template.restored", "templates", &id, ip, userAgent)
	}
	return err
}

func (s *TemplateService) ForceDeleteTemplate(id uuid.UUID, ip, userAgent string) error {
	err := s.templateRepo.ForceDelete(id)
	if err == nil {
		s.auditLog(nil, "template.force_deleted", "templates", &id, ip, userAgent)
	}
	return err
}
