package services

import (
	"errors"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/repositories"
)

type CompanySettingService struct {
	repo      repositories.CompanySettingRepository
	auditRepo repositories.AuditLogRepository
}

func NewCompanySettingService(repo repositories.CompanySettingRepository, auditRepo repositories.AuditLogRepository) *CompanySettingService {
	return &CompanySettingService{repo: repo, auditRepo: auditRepo}
}

func (s *CompanySettingService) GetSettings() (*models.CompanySetting, error) {
	setting, err := s.repo.Get()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default if not found (though seed should provide it)
			return &models.CompanySetting{Name: "Wedding Company"}, nil
		}
		return nil, err
	}
	return setting, nil
}

func (s *CompanySettingService) UpdateSettings(
	req models.UpdateCompanySettingRequest,
	faviconURL, logoLongURL, logoSquareURL *string,
	userID uuid.UUID, ip, userAgent string,
) (*models.CompanySetting, error) {
	setting, err := s.repo.Get()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			setting = &models.CompanySetting{}
		} else {
			return nil, err
		}
	}

	setting.Name = req.Name
	setting.Address = req.Address
	setting.Description = req.Description

	if faviconURL != nil {
		setting.FaviconURL = *faviconURL
	}
	if logoLongURL != nil {
		setting.LogoLongURL = *logoLongURL
	}
	if logoSquareURL != nil {
		setting.LogoSquareURL = *logoSquareURL
	}

	if setting.ID == uuid.Nil {
		if err := s.repo.Create(setting); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(setting); err != nil {
			return nil, err
		}
	}

	s.auditRepo.Create(&models.AuditLog{
		UserID:       &userID,
		Action:       "company_setting.updated",
		ResourceType: "company_settings",
		ResourceID:   &setting.ID,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
	})

	return setting, nil
}
