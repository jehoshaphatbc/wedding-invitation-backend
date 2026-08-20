package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type EmailVerificationTokenRepository interface {
	Create(token *models.EmailVerificationToken) error
	FindByTokenHash(tokenHash string) (*models.EmailVerificationToken, error)
	MarkAsVerified(id uuid.UUID) error
	DeleteExpired() error
}

type emailVerificationTokenRepository struct {
	db *gorm.DB
}

func NewEmailVerificationTokenRepository(db *gorm.DB) EmailVerificationTokenRepository {
	return &emailVerificationTokenRepository{db: db}
}

func (r *emailVerificationTokenRepository) Create(token *models.EmailVerificationToken) error {
	return r.db.Create(token).Error
}

func (r *emailVerificationTokenRepository) FindByTokenHash(tokenHash string) (*models.EmailVerificationToken, error) {
	var token models.EmailVerificationToken
	err := r.db.Where("token_hash = ? AND verified_at IS NULL AND expires_at > ?", tokenHash, time.Now()).First(&token).Error
	return &token, err
}

func (r *emailVerificationTokenRepository) MarkAsVerified(id uuid.UUID) error {
	return r.db.Model(&models.EmailVerificationToken{}).Where("id = ?", id).Update("verified_at", time.Now()).Error
}

func (r *emailVerificationTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.EmailVerificationToken{}).Error
}
