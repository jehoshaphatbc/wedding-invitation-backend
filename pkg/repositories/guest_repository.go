package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type GuestRepository interface {
	FindByInvitationID(invitationID uuid.UUID) ([]models.Guest, error)
	FindByIDAndInvitationID(id, invitationID uuid.UUID) (*models.Guest, error)
	Create(guest *models.Guest) error
	BulkCreate(guests []models.Guest) error
	Update(guest *models.Guest) error
	Delete(id, invitationID uuid.UUID) error
}

type guestRepository struct {
	db *gorm.DB
}

func NewGuestRepository(db *gorm.DB) GuestRepository {
	return &guestRepository{db: db}
}

func (r *guestRepository) FindByInvitationID(invitationID uuid.UUID) ([]models.Guest, error) {
	var guests []models.Guest
	err := r.db.Where("invitation_id = ?", invitationID).Order("created_at ASC").Find(&guests).Error
	return guests, err
}

func (r *guestRepository) FindByIDAndInvitationID(id, invitationID uuid.UUID) (*models.Guest, error) {
	var guest models.Guest
	err := r.db.Where("id = ? AND invitation_id = ?", id, invitationID).First(&guest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &guest, nil
}

func (r *guestRepository) Create(guest *models.Guest) error {
	return r.db.Create(guest).Error
}

func (r *guestRepository) BulkCreate(guests []models.Guest) error {
	if len(guests) == 0 {
		return nil
	}
	return r.db.Create(&guests).Error
}

func (r *guestRepository) Update(guest *models.Guest) error {
	return r.db.Save(guest).Error
}

func (r *guestRepository) Delete(id, invitationID uuid.UUID) error {
	return r.db.Where("id = ? AND invitation_id = ?", id, invitationID).Delete(&models.Guest{}).Error
}
