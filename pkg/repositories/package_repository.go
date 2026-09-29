package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type PackageRepository interface {
	Create(pkg *models.Package) error
	FindAll() ([]models.Package, error)
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

func (r *packageRepository) FindAll() ([]models.Package, error) {
	var packages []models.Package
	err := r.db.Order("created_at DESC").Find(&packages).Error
	return packages, err
}

func (r *packageRepository) FindByID(id uuid.UUID) (*models.Package, error) {
	var pkg models.Package
	err := r.db.Where("id = ?", id).First(&pkg).Error
	return &pkg, err
}

func (r *packageRepository) Update(pkg *models.Package) error {
	return r.db.Save(pkg).Error
}

func (r *packageRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Package{}).Error
}
