package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type TemplateRepository interface {
	Create(template *models.Template) error
	FindAll(page, perPage int, search, category string, isActive *bool, isTrashed bool) ([]models.Template, int64, error)
	FindByID(id uuid.UUID) (*models.Template, error)
	Update(template *models.Template) error
	Delete(id uuid.UUID) error
	FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Template, int64, error)
	Restore(id uuid.UUID) error
	ForceDelete(id uuid.UUID) error
}

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(template *models.Template) error {
	return r.db.Create(template).Error
}

func (r *templateRepository) FindAll(page, perPage int, search, category string, isActive *bool, isTrashed bool) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	query := r.db.Model(&models.Template{})
	if isTrashed {
		query = r.db.Unscoped().Model(&models.Template{}).Where("deleted_at IS NOT NULL")
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR nuxt_component ILIKE ? OR category ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if category != "" {
		query = query.Where("LOWER(category) = LOWER(?)", category)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if perPage > 0 {
		offset := (page - 1) * perPage
		query = query.Offset(offset).Limit(perPage)
	}

	err := query.Order("created_at desc").Find(&templates).Error
	return templates, total, err
}

func (r *templateRepository) FindByID(id uuid.UUID) (*models.Template, error) {
	var template models.Template
	err := r.db.Where("id = ?", id).First(&template).Error
	return &template, err
}

func (r *templateRepository) Update(template *models.Template) error {
	return r.db.Save(template).Error
}

func (r *templateRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Template{}).Error
}

func (r *templateRepository) FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	query := r.db.Unscoped().Where("deleted_at IS NOT NULL")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR nuxt_component ILIKE ? OR category ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	query.Model(&models.Template{}).Count(&total)

	if sort == "" {
		sort = "deleted_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&templates).Error

	return templates, total, err
}

func (r *templateRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().Model(&models.Template{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *templateRepository) ForceDelete(id uuid.UUID) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.Template{}).Error
}
