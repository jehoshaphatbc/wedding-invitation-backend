package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindAll(page, perPage int, search, status, role, sort, order string) ([]models.User, int64, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	FindTrashedAll(page, perPage int, search, sort, order string) ([]models.User, int64, error)
	Restore(id uuid.UUID) error
	ForceDelete(id uuid.UUID) error
	Count() (int64, error)
	AssignRoles(userID uuid.UUID, roleIDs []uuid.UUID) error
	UpdateLastLogin(userID uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles.Permissions").Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles.Permissions").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) FindAll(page, perPage int, search, status, role, sort, order string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if role != "" {
		query = query.Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("roles.name = ?", role)
	}

	query.Count(&total)

	if sort == "" {
		sort = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Preload("Roles.Permissions").Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&users).Error

	return users, total, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Model(user).Select("*").Omit("Roles", "ClientProfile").Updates(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

func (r *userRepository) AssignRoles(userID uuid.UUID, roleIDs []uuid.UUID) error {
	var userRoles []models.UserRole
	for _, roleID := range roleIDs {
		userRoles = append(userRoles, models.UserRole{
			UserID:    userID,
			RoleID:    roleID,
			CreatedAt: time.Now(),
		})
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("user_id = ?", userID).Delete(&models.UserRole{})
		return tx.Create(&userRoles).Error
	})
}

func (r *userRepository) UpdateLastLogin(userID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *userRepository) FindTrashedAll(page, perPage int, search, sort, order string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Unscoped().Where("deleted_at IS NOT NULL")

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Model(&models.User{}).Count(&total)

	if sort == "" {
		sort = "deleted_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Preload("Roles.Permissions").Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&users).Error

	return users, total, err
}

func (r *userRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().Model(&models.User{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *userRepository) ForceDelete(id uuid.UUID) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.User{}).Error
}
