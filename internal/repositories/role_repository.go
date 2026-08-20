package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type RoleRepository interface {
	Create(role *models.Role) error
	FindByID(id uuid.UUID) (*models.Role, error)
	FindByName(name string) (*models.Role, error)
	FindAll() ([]models.Role, error)
	Update(role *models.Role) error
	Delete(id uuid.UUID) error
	AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	Count() (int64, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) FindByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	return &role, err
}

func (r *roleRepository) FindByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *roleRepository) FindAll() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").Order("created_at ASC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Role{}).Error
}

func (r *roleRepository) AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	var rolePermissions []models.RolePermission
	for _, permID := range permissionIDs {
		rolePermissions = append(rolePermissions, models.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		})
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{})
		if len(rolePermissions) > 0 {
			return tx.Create(&rolePermissions).Error
		}
		return nil
	})
}

func (r *roleRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Role{}).Count(&count).Error
	return count, err
}
