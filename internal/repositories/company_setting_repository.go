package repositories

import (
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type CompanySettingRepository interface {
	Get() (*models.CompanySetting, error)
	Update(setting *models.CompanySetting) error
	Create(setting *models.CompanySetting) error
}

type companySettingRepository struct {
	db *gorm.DB
}

func NewCompanySettingRepository(db *gorm.DB) CompanySettingRepository {
	return &companySettingRepository{db: db}
}

func (r *companySettingRepository) Get() (*models.CompanySetting, error) {
	var setting models.CompanySetting
	err := r.db.First(&setting).Error
	return &setting, err
}

func (r *companySettingRepository) Update(setting *models.CompanySetting) error {
	return r.db.Save(setting).Error
}

func (r *companySettingRepository) Create(setting *models.CompanySetting) error {
	return r.db.Create(setting).Error
}
