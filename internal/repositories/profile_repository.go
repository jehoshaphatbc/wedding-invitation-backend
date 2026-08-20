package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

type ProfileRepository interface {
	Create(profile *models.ClientProfile) error
	FindByUserID(userID uuid.UUID) (*models.ClientProfile, error)
	Update(profile *models.ClientProfile) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) Create(profile *models.ClientProfile) error {
	return r.db.Create(profile).Error
}

func (r *profileRepository) FindByUserID(userID uuid.UUID) (*models.ClientProfile, error) {
	var profile models.ClientProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

func (r *profileRepository) Update(profile *models.ClientProfile) error {
	return r.db.Save(profile).Error
}
