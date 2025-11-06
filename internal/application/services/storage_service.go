package services

import (
	"context"
	"io"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
)

// StorageService implements business logic for file storage operations
type StorageService struct {
	repo          ports.StorageRepository
	defaultBucket string
}

// NewStorageService creates a new storage service
func NewStorageService(repo ports.StorageRepository, defaultBucket string) *StorageService {
	return &StorageService{
		repo:          repo,
		defaultBucket: defaultBucket,
	}
}

// UploadFile uploads a file to the specified bucket with the given filename
func (s *StorageService) UploadFile(ctx context.Context, bucket, fileName string, reader io.Reader, size int64, contentType string) (*entities.FileMetadata, error) {
	// Use default bucket if not specified
	if bucket == "" {
		bucket = s.defaultBucket
	}

	// Validate filename
	if fileName == "" {
		return nil, domainErrors.NewValidationError("file_name", "filename is required")
	}

	// Ensure bucket exists
	err := s.repo.EnsureBucket(ctx, bucket)
	if err != nil {
		return nil, err
	}

	// Store file using repository
	metadata, err := s.repo.Store(ctx, bucket, fileName, reader, size, contentType)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// DeleteFile deletes a file from the specified bucket
func (s *StorageService) DeleteFile(ctx context.Context, bucket, fileName string) error {
	// Use default bucket if not specified
	if bucket == "" {
		bucket = s.defaultBucket
	}

	// Delete file using repository
	err := s.repo.Remove(ctx, bucket, fileName)
	if err != nil {
		return err
	}

	return nil
}

// GetFileURL generates a public URL for the file
func (s *StorageService) GetFileURL(ctx context.Context, bucket, fileName string) (string, error) {
	// Use default bucket if not specified
	if bucket == "" {
		bucket = s.defaultBucket
	}

	// Get URL from repository
	url, err := s.repo.GetURL(ctx, bucket, fileName)
	if err != nil {
		return "", err
	}

	return url, nil
}
