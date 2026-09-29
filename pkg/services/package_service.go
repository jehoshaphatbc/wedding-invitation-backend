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
}

func NewPackageService(packageRepo repositories.PackageRepository) *PackageService {
	return &PackageService{packageRepo: packageRepo}
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

func (s *PackageService) DeletePackage(id uuid.UUID) error {
	_, err := s.GetPackageByID(id)
	if err != nil {
		return err
	}

	return s.packageRepo.Delete(id)
}
