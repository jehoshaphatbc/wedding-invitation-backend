package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type PermissionRepository interface {
	Create(permission *models.Permission) error
	FindAll() ([]models.Permission, error)
	FindByID(id uuid.UUID) (*models.Permission, error)
	Update(permission *models.Permission) error
	Delete(id uuid.UUID) error
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

func (r *permissionRepository) FindByID(id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("id = ?", id).First(&permission).Error
	return &permission, err
}

func (r *permissionRepository) Update(permission *models.Permission) error {
	return r.db.Model(permission).Where("id = ?", permission.ID).Updates(map[string]interface{}{
		"display_name": permission.DisplayName,
		"description":  permission.Description,
	}).Error
}

func (r *permissionRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("permission_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.Permission{}).Error
	})
}

func (r *permissionRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Permission{}).Count(&count).Error
	return count, err
}

