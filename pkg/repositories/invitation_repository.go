package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type InvitationRepository interface {
	Create(invitation *models.Invitation) error
	FindByOrderID(orderID uuid.UUID) (*models.Invitation, error)
	FindByID(id uuid.UUID) (*models.Invitation, error)
	Update(invitation *models.Invitation) error
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(invitation *models.Invitation) error {
	return r.db.Create(invitation).Error
}

func (r *invitationRepository) FindByOrderID(orderID uuid.UUID) (*models.Invitation, error) {
	var invitation models.Invitation
	err := r.db.Where("order_id = ?", orderID).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) FindByID(id uuid.UUID) (*models.Invitation, error) {
	var invitation models.Invitation
	err := r.db.Where("id = ?", id).First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) Update(invitation *models.Invitation) error {
	return r.db.Save(invitation).Error
}
