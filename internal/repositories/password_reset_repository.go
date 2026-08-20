package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type PasswordResetTokenRepository interface {
	Create(token *models.PasswordResetToken) error
	FindByTokenHash(tokenHash string) (*models.PasswordResetToken, error)
	MarkAsUsed(id uuid.UUID) error
	DeleteExpired() error
}

type passwordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(token *models.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *passwordResetTokenRepository) FindByTokenHash(tokenHash string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	err := r.db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).First(&token).Error
	return &token, err
}

func (r *passwordResetTokenRepository) MarkAsUsed(id uuid.UUID) error {
	return r.db.Model(&models.PasswordResetToken{}).Where("id = ?", id).Update("used_at", time.Now()).Error
}

func (r *passwordResetTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.PasswordResetToken{}).Error
}
