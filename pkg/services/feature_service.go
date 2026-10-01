package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type FeatureService struct {
	featureRepo repositories.FeatureRepository
	auditRepo   repositories.AuditLogRepository
}

func NewFeatureService(featureRepo repositories.FeatureRepository, auditRepo repositories.AuditLogRepository) *FeatureService {
	return &FeatureService{
		featureRepo: featureRepo,
		auditRepo:   auditRepo,
	}
}

func (s *FeatureService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
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

func (s *FeatureService) CreateFeature(req models.CreateFeatureRequest, ip, userAgent string) (*models.Feature, error) {
	existing, err := s.featureRepo.FindByKey(req.FeatureKey)
	if err == nil && existing != nil {
		return nil, errors.New("feature_key already exists")
	}

	feature := &models.Feature{
		FeatureKey:   req.FeatureKey,
		FeatureName:  req.FeatureName,
		InputType:    req.InputType,
		DefaultValue: req.DefaultValue,
	}

	if err := s.featureRepo.Create(feature); err != nil {
		return nil, err
	}

	s.auditLog(nil, "feature.created", "features", &feature.ID, ip, userAgent)
	return feature, nil
}

func (s *FeatureService) GetAllFeatures(page, perPage int, search, inputType, sort, order string) ([]models.Feature, int64, error) {
	return s.featureRepo.FindAll(page, perPage, search, inputType, sort, order)
}

func (s *FeatureService) GetFeatureByID(id uuid.UUID) (*models.Feature, error) {
	feature, err := s.featureRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("feature not found")
		}
		return nil, err
	}
	return feature, nil
}

func (s *FeatureService) UpdateFeature(id uuid.UUID, req models.UpdateFeatureRequest, ip, userAgent string) (*models.Feature, error) {
	feature, err := s.GetFeatureByID(id)
	if err != nil {
		return nil, err
	}

	if req.FeatureKey != nil && *req.FeatureKey != "" && *req.FeatureKey != feature.FeatureKey {
		existing, err := s.featureRepo.FindByKey(*req.FeatureKey)
		if err == nil && existing != nil && existing.ID != id {
			return nil, errors.New("feature_key already exists")
		}
		feature.FeatureKey = *req.FeatureKey
	}
	if req.FeatureName != nil && *req.FeatureName != "" {
		feature.FeatureName = *req.FeatureName
	}
	if req.InputType != nil && *req.InputType != "" {
		feature.InputType = *req.InputType
	}
	if req.DefaultValue != nil {
		feature.DefaultValue = *req.DefaultValue
	}

	if err := s.featureRepo.Update(feature); err != nil {
		return nil, err
	}

	s.auditLog(nil, "feature.updated", "features", &feature.ID, ip, userAgent)
	return s.GetFeatureByID(id)
}

func (s *FeatureService) DeleteFeature(id uuid.UUID, ip, userAgent string) error {
	feature, err := s.GetFeatureByID(id)
	if err != nil {
		return err
	}

	if err := s.featureRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "feature.deleted", "features", &feature.ID, ip, userAgent)
	return nil
}
