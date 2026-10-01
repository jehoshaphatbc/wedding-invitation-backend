package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)

type OrderRepository interface {
	Create(order *models.Order) error
	FindAll(page, perPage int, search, status string, isTrashed bool) ([]models.Order, int64, error)
	FindByID(id uuid.UUID) (*models.Order, error)
	FindByInvoiceNumber(invoice string) (*models.Order, error)
	FindByFormToken(formToken string) (*models.Order, error)
	Update(order *models.Order) error
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) error
	ForceDelete(id uuid.UUID) error
	FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Order, int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindAll(page, perPage int, search, status string, isTrashed bool) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Preload("Client").Preload("Package")
	if isTrashed {
		query = r.db.Unscoped().Model(&models.Order{}).Where("deleted_at IS NOT NULL").Preload("Client").Preload("Package")
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("LEFT JOIN clients ON clients.id = orders.client_id").
			Where("orders.invoice_number ILIKE ? OR clients.name ILIKE ? OR clients.email ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if status != "" {
		query = query.Where("orders.status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if perPage > 0 {
		offset := (page - 1) * perPage
		query = query.Offset(offset).Limit(perPage)
	}

	err := query.Order("orders.created_at desc").Find(&orders).Error
	return orders, total, err
}

func (r *orderRepository) FindByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Client").Preload("Package").Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByInvoiceNumber(invoice string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Client").Preload("Package").Where("invoice_number = ?", invoice).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByFormToken(formToken string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Client").Preload("Package").Where("form_token = ?", formToken).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Order{}).Error
}

func (r *orderRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().Model(&models.Order{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *orderRepository) ForceDelete(id uuid.UUID) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.Order{}).Error
}

func (r *orderRepository) FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Unscoped().Where("deleted_at IS NOT NULL").Preload("Client").Preload("Package")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("LEFT JOIN clients ON clients.id = orders.client_id").
			Where("orders.invoice_number ILIKE ? OR clients.name ILIKE ? OR clients.email ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	query.Model(&models.Order{}).Count(&total)

	if sort == "" {
		sort = "deleted_at"
	}
	if order == "" {
		order = "desc"
	}

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Order(sort + " " + order).Find(&orders).Error
	return orders, total, err
}
