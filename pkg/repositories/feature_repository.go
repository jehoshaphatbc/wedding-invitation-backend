package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type FeatureRepository interface {
	Create(feature *models.Feature) error
	FindAll(page, perPage int, search, inputType, sort, order string) ([]models.Feature, int64, error)
	FindByID(id uuid.UUID) (*models.Feature, error)
	FindByKey(key string) (*models.Feature, error)
	Update(feature *models.Feature) error
	Delete(id uuid.UUID) error
	Count() (int64, error)
}

type featureRepository struct {
	db *gorm.DB
}

func NewFeatureRepository(db *gorm.DB) FeatureRepository {
	return &featureRepository{db: db}
}

func (r *featureRepository) Create(feature *models.Feature) error {
	return r.db.Create(feature).Error
}

func (r *featureRepository) FindAll(page, perPage int, search, inputType, sort, order string) ([]models.Feature, int64, error) {
	var features []models.Feature
	var total int64

	query := r.db.Model(&models.Feature{})

	if search != "" {
		query = query.Where("feature_name ILIKE ? OR feature_key ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if inputType != "" {
		query = query.Where("input_type = ?", inputType)
	}

	query.Count(&total)

	if sort == "" {
		sort = "created_at"
	}
	if order == "" {
		order = "asc"
	}

	if perPage > 0 {
		offset := (page - 1) * perPage
		query = query.Offset(offset).Limit(perPage)
	}

	err := query.Order(sort + " " + order).Find(&features).Error
	return features, total, err
}

func (r *featureRepository) FindByID(id uuid.UUID) (*models.Feature, error) {
	var feature models.Feature
	err := r.db.Where("id = ?", id).First(&feature).Error
	return &feature, err
}

func (r *featureRepository) FindByKey(key string) (*models.Feature, error) {
	var feature models.Feature
	err := r.db.Where("feature_key = ?", key).First(&feature).Error
	return &feature, err
}

func (r *featureRepository) Update(feature *models.Feature) error {
	return r.db.Model(feature).Where("id = ?", feature.ID).Updates(map[string]interface{}{
		"feature_key":   feature.FeatureKey,
		"feature_name":  feature.FeatureName,
		"input_type":    feature.InputType,
		"default_value": feature.DefaultValue,
	}).Error
}

func (r *featureRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Feature{}).Error
}

func (r *featureRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Feature{}).Count(&count).Error
	return count, err
}
