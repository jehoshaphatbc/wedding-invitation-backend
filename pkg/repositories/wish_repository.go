package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type WishRepository interface {
	Create(wish *models.Wish) error
	FindByInvitationID(invitationID uuid.UUID) ([]models.Wish, error)
}

type wishRepository struct {
	db *gorm.DB
}

func NewWishRepository(db *gorm.DB) WishRepository {
	return &wishRepository{db: db}
}

func (r *wishRepository) Create(wish *models.Wish) error {
	return r.db.Create(wish).Error
}

func (r *wishRepository) FindByInvitationID(invitationID uuid.UUID) ([]models.Wish, error) {
	var wishes []models.Wish
	err := r.db.Where("invitation_id = ?", invitationID).Order("created_at DESC").Find(&wishes).Error
	return wishes, err
}

