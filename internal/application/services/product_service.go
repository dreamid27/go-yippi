package services

import (
	"context"
	"strings"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
)

// ProductService handles business logic for products
type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(repo ports.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *entities.Product) error {
	// Validate required fields
	if strings.TrimSpace(product.SKU) == "" {
		return domainErrors.NewValidationError("sku", "SKU is required")
	}
	if strings.TrimSpace(product.Slug) == "" {
		return domainErrors.NewValidationError("slug", "Slug is required")
	}
	if strings.TrimSpace(product.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}
	if product.Price <= 0 {
		return domainErrors.NewValidationError("price", "Price must be greater than 0")
	}

	// Validate status
	if !product.IsValid() {
		return domainErrors.NewValidationError("status", "Invalid product status")
	}

	// Validate dimensions for courier calculation
	if product.Weight < 0 {
		return domainErrors.NewValidationError("weight", "Weight cannot be negative")
	}

	return s.repo.Create(ctx, product)
}

func (s *ProductService) GetProduct(ctx context.Context, id int) (*entities.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	if strings.TrimSpace(sku) == "" {
		return nil, domainErrors.NewValidationError("sku", "SKU is required")
	}
	return s.repo.GetBySKU(ctx, sku)
}

func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (*entities.Product, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, domainErrors.NewValidationError("slug", "Slug is required")
	}
	return s.repo.GetBySlug(ctx, slug)
}

func (s *ProductService) ListProducts(ctx context.Context) ([]*entities.Product, error) {
	return s.repo.List(ctx)
}

func (s *ProductService) ListPublishedProducts(ctx context.Context) ([]*entities.Product, error) {
	return s.repo.ListByStatus(ctx, entities.ProductStatusPublished)
}

func (s *ProductService) ListProductsByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error) {
	return s.repo.ListByStatus(ctx, status)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *entities.Product) error {
	// Validate required fields
	if strings.TrimSpace(product.SKU) == "" {
		return domainErrors.NewValidationError("sku", "SKU is required")
	}
	if strings.TrimSpace(product.Slug) == "" {
		return domainErrors.NewValidationError("slug", "Slug is required")
	}
	if strings.TrimSpace(product.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}
	if product.Price <= 0 {
		return domainErrors.NewValidationError("price", "Price must be greater than 0")
	}

	// Validate status
	if !product.IsValid() {
		return domainErrors.NewValidationError("status", "Invalid product status")
	}

	return s.repo.Update(ctx, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) PublishProduct(ctx context.Context, id int) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Business rule: can only publish draft products
	if product.Status != entities.ProductStatusDraft {
		return domainErrors.NewValidationError("status", "Only draft products can be published")
	}

	product.Status = entities.ProductStatusPublished
	return s.repo.Update(ctx, product)
}

func (s *ProductService) ArchiveProduct(ctx context.Context, id int) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	product.Status = entities.ProductStatusArchived
	return s.repo.Update(ctx, product)
}
