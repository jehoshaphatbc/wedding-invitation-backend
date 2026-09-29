package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type RoleRepository interface {
	Create(role *models.Role) error
	FindByID(id uuid.UUID) (*models.Role, error)
	FindByName(name string) (*models.Role, error)
	FindAll(page, perPage int, search, sort, order string) ([]models.Role, int64, error)
	Update(role *models.Role) error
	Delete(id uuid.UUID) error
	FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Role, int64, error)
	Restore(id uuid.UUID) error
	ForceDelete(id uuid.UUID) error
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
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *roleRepository) FindAll(page, perPage int, search, sort, order string) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := r.db.Model(&models.Role{})

	if search != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	if sort == "" {
		sort = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Preload("Permissions").Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&roles).Error

	return roles, total, err
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

func (r *roleRepository) FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := r.db.Unscoped().Where("deleted_at IS NOT NULL")

	if search != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Model(&models.Role{}).Count(&total)

	if sort == "" {
		sort = "deleted_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Preload("Permissions").Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&roles).Error

	return roles, total, err
}

func (r *roleRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().Model(&models.Role{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *roleRepository) ForceDelete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete dependent records
		if err := tx.Where("role_id = ?", id).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		// Delete the role itself
		return tx.Unscoped().Where("id = ?", id).Delete(&models.Role{}).Error
	})
}
