package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type AuditLogRepository interface {
	Create(log *models.AuditLog) error
	FindByUserID(userID uuid.UUID, page, perPage int) ([]models.AuditLog, int64, error)
	FindAll(page, perPage int) ([]models.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) FindByUserID(userID uuid.UUID, page, perPage int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *auditLogRepository) FindAll(page, perPage int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	r.db.Model(&models.AuditLog{}).Count(&total)

	offset := (page - 1) * perPage
	err := r.db.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}
