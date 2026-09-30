package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type PackageRequest struct {
	Name           string                `json:"name" binding:"required"`
	Price          float64               `json:"price" binding:"required"`
	FeaturesConfig models.FeaturesConfig `json:"features_config" binding:"required"`
}

type PackageService struct {
	packageRepo repositories.PackageRepository
	auditRepo   repositories.AuditLogRepository
}

func NewPackageService(packageRepo repositories.PackageRepository, auditRepo repositories.AuditLogRepository) *PackageService {
	return &PackageService{
		packageRepo: packageRepo,
		auditRepo:   auditRepo,
	}
}

func (s *PackageService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
	if s.auditRepo == nil {
		return
	}
	s.auditRepo.Create(&models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
	})
}

func (s *PackageService) CreatePackage(req PackageRequest) (*models.Package, error) {
	pkg := &models.Package{
		Name:           req.Name,
		Price:          req.Price,
		FeaturesConfig: req.FeaturesConfig,
	}

	if err := s.packageRepo.Create(pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

func (s *PackageService) GetAllPackages(page, perPage int, search, sort, order string) ([]models.Package, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.packageRepo.FindAll(page, perPage, search, sort, order)
}

func (s *PackageService) GetPackageByID(id uuid.UUID) (*models.Package, error) {
	pkg, err := s.packageRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("package not found")
		}
		return nil, err
	}
	return pkg, nil
}

func (s *PackageService) UpdatePackage(id uuid.UUID, req PackageRequest) (*models.Package, error) {
	pkg, err := s.GetPackageByID(id)
	if err != nil {
		return nil, err
	}

	pkg.Name = req.Name
	pkg.Price = req.Price
	pkg.FeaturesConfig = req.FeaturesConfig

	if err := s.packageRepo.Update(pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

func (s *PackageService) DeletePackage(id uuid.UUID, ip, userAgent string) error {
	pkg, err := s.GetPackageByID(id)
	if err != nil {
		return err
	}

	if err := s.packageRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "package.deleted", "packages", &pkg.ID, ip, userAgent)
	return nil
}

func (s *PackageService) GetTrashedPackages(page, perPage int, search, sort, order string) ([]models.Package, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.packageRepo.FindTrashedAll(page, perPage, search, sort, order)
}

func (s *PackageService) RestorePackage(id uuid.UUID, ip, userAgent string) error {
	err := s.packageRepo.Restore(id)
	if err == nil {
		s.auditLog(nil, "package.restored", "packages", &id, ip, userAgent)
	}
	return err
}

func (s *PackageService) ForceDeletePackage(id uuid.UUID, ip, userAgent string) error {
	err := s.packageRepo.ForceDelete(id)
	if err == nil {
		s.auditLog(nil, "package.force_deleted", "packages", &id, ip, userAgent)
	}
	return err
}
