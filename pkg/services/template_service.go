package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type TemplateRequest struct {
	Name          string  `json:"name" binding:"required"`
	NuxtComponent string  `json:"nuxt_component" binding:"required"`
	ThumbnailURL  *string `json:"thumbnail_url"`
}

type TemplateService struct {
	templateRepo repositories.TemplateRepository
}

func NewTemplateService(templateRepo repositories.TemplateRepository) *TemplateService {
	return &TemplateService{templateRepo: templateRepo}
}

func (s *TemplateService) CreateTemplate(req TemplateRequest) (*models.Template, error) {
	template := &models.Template{
		Name:          req.Name,
		NuxtComponent: req.NuxtComponent,
		ThumbnailURL:  req.ThumbnailURL,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *TemplateService) GetAllTemplates() ([]models.Template, error) {
	return s.templateRepo.FindAll()
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
	template.ThumbnailURL = req.ThumbnailURL

	if err := s.templateRepo.Update(template); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *TemplateService) DeleteTemplate(id uuid.UUID) error {
	_, err := s.GetTemplateByID(id)
	if err != nil {
		return err
	}

	return s.templateRepo.Delete(id)
}
