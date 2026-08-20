package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByTokenHash(tokenHash string) (*models.RefreshToken, error)
	FindByUserID(userID uuid.UUID) ([]models.RefreshToken, error)
	Revoke(id uuid.UUID) error
	RevokeAllByUserID(userID uuid.UUID) error
	DeleteExpired() error
	UpdateLastUsed(id uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, time.Now()).First(&token).Error
	return &token, err
}

func (r *refreshTokenRepository) FindByUserID(userID uuid.UUID) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := r.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).Find(&tokens).Error
	return tokens, err
}

func (r *refreshTokenRepository) Revoke(id uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).Where("id = ?", id).Update("revoked_at", time.Now()).Error
}

func (r *refreshTokenRepository) RevokeAllByUserID(userID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", time.Now()).Error
}

func (r *refreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ? OR revoked_at IS NOT NULL", time.Now()).Delete(&models.RefreshToken{}).Error
}

func (r *refreshTokenRepository) UpdateLastUsed(id uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).Where("id = ?", id).Update("last_used_at", time.Now()).Error
}
