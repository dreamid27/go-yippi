package services

import (
	"context"
	"strings"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/google/uuid"
)

// BrandService handles business logic for brands
type BrandService struct {
	repo ports.BrandRepository
}

func NewBrandService(repo ports.BrandRepository) *BrandService {
	return &BrandService{repo: repo}
}

func (s *BrandService) CreateBrand(ctx context.Context, brand *entities.Brand) error {
	// Validate required fields
	if strings.TrimSpace(brand.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}

	// Validate name length (must not exceed 255 characters)
	if len(brand.Name) > 255 {
		return domainErrors.NewValidationError("name", "Name must not exceed 255 characters")
	}

	return s.repo.Create(ctx, brand)
}

func (s *BrandService) GetBrand(ctx context.Context, id uuid.UUID) (*entities.Brand, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BrandService) GetBrandByName(ctx context.Context, name string) (*entities.Brand, error) {
	if strings.TrimSpace(name) == "" {
		return nil, domainErrors.NewValidationError("name", "Name is required")
	}
	return s.repo.GetByName(ctx, name)
}

func (s *BrandService) ListBrands(ctx context.Context) ([]*entities.Brand, error) {
	return s.repo.List(ctx)
}

func (s *BrandService) UpdateBrand(ctx context.Context, brand *entities.Brand) error {
	// Validate required fields
	if strings.TrimSpace(brand.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}

	// Validate name length (must not exceed 255 characters)
	if len(brand.Name) > 255 {
		return domainErrors.NewValidationError("name", "Name must not exceed 255 characters")
	}

	return s.repo.Update(ctx, brand)
}

func (s *BrandService) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
