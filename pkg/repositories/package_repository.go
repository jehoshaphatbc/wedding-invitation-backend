package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type PackageRepository interface {
	Create(pkg *models.Package) error
	FindAll(page, perPage int, search, sort, order string) ([]models.Package, int64, error)
	FindByID(id uuid.UUID) (*models.Package, error)
	Update(pkg *models.Package) error
	Delete(id uuid.UUID) error
}

type packageRepository struct {
	db *gorm.DB
}

func NewPackageRepository(db *gorm.DB) PackageRepository {
	return &packageRepository{db: db}
}

func (r *packageRepository) Create(pkg *models.Package) error {
	return r.db.Create(pkg).Error
}

func (r *packageRepository) FindAll(page, perPage int, search, sort, order string) ([]models.Package, int64, error) {
	var packages []models.Package
	var total int64

	query := r.db.Model(&models.Package{})

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	query.Count(&total)

	if sort == "" {
		sort = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&packages).Error
	return packages, total, err
}

func (r *packageRepository) FindByID(id uuid.UUID) (*models.Package, error) {
	var pkg models.Package
	err := r.db.Where("id = ?", id).First(&pkg).Error
	return &pkg, err
}

func (r *packageRepository) Update(pkg *models.Package) error {
	return r.db.Model(pkg).Where("id = ?", pkg.ID).Updates(map[string]interface{}{
		"name":            pkg.Name,
		"price":           pkg.Price,
		"features_config": pkg.FeaturesConfig,
	}).Error
}

func (r *packageRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Package{}).Error
}
