package repositories

import (
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type PermissionRepository interface {
	Create(permission *models.Permission) error
	FindAll() ([]models.Permission, error)
	Count() (int64, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) FindAll() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Order("name ASC").Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Permission{}).Count(&count).Error
	return count, err
}
