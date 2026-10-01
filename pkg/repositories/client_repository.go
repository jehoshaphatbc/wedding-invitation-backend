package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type ClientRepository interface {
	Create(client *models.Client) error
	FindAll(page, perPage int, search string, isTrashed bool) ([]models.Client, int64, error)
	FindByID(id uuid.UUID) (*models.Client, error)
	FindByEmail(email string) (*models.Client, error)
	Update(client *models.Client) error
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) error
	ForceDelete(id uuid.UUID) error
	FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Client, int64, error)
}

type clientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) ClientRepository {
	return &clientRepository{db: db}
}

func (r *clientRepository) Create(client *models.Client) error {
	return r.db.Create(client).Error
}

func (r *clientRepository) FindAll(page, perPage int, search string, isTrashed bool) ([]models.Client, int64, error) {
	var clients []models.Client
	var total int64

	query := r.db.Model(&models.Client{})
	if isTrashed {
		query = r.db.Unscoped().Model(&models.Client{}).Where("deleted_at IS NOT NULL")
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR whatsapp ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if perPage > 0 {
		offset := (page - 1) * perPage
		query = query.Offset(offset).Limit(perPage)
	}

	err := query.Preload("Orders.Package").Order("created_at desc").Find(&clients).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range clients {
		if clients[i].Orders == nil {
			clients[i].Orders = []models.Order{}
		}
	}

	return clients, total, nil
}

func (r *clientRepository) FindByID(id uuid.UUID) (*models.Client, error) {
	var client models.Client
	err := r.db.Preload("Orders.Package").Where("id = ?", id).First(&client).Error
	if err != nil {
		return nil, err
	}
	if client.Orders == nil {
		client.Orders = []models.Order{}
	}
	return &client, nil
}

func (r *clientRepository) FindByEmail(email string) (*models.Client, error) {
	var client models.Client
	err := r.db.Where("email = ?", email).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &client, nil
}

func (r *clientRepository) Update(client *models.Client) error {
	return r.db.Save(client).Error
}

func (r *clientRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Client{}).Error
}

func (r *clientRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().Model(&models.Client{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *clientRepository) ForceDelete(id uuid.UUID) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.Client{}).Error
}

func (r *clientRepository) FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Client, int64, error) {
	var clients []models.Client
	var total int64

	query := r.db.Unscoped().Where("deleted_at IS NOT NULL")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR whatsapp ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	query.Model(&models.Client{}).Count(&total)

	if sort == "" {
		sort = "deleted_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Preload("Orders.Package").Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&clients).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range clients {
		if clients[i].Orders == nil {
			clients[i].Orders = []models.Order{}
		}
	}

	return clients, total, nil
}
