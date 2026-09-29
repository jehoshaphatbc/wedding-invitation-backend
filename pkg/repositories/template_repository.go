package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type TemplateRepository interface {
	Create(template *models.Template) error
	FindAll(page, perPage int, search, sort, order string) ([]models.Template, int64, error)
	FindByID(id uuid.UUID) (*models.Template, error)
	Update(template *models.Template) error
	Delete(id uuid.UUID) error
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

func (r *templateRepository) FindAll(page, perPage int, search, sort, order string) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	query := r.db.Model(&models.Template{})

	if search != "" {
		query = query.Where("name ILIKE ? OR nuxt_component ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	if sort == "" {
		sort = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&templates).Error
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
