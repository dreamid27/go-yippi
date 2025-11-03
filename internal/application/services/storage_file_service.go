package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/google/uuid"
)

// StorageFileService handles business logic for file storage
type StorageFileService struct {
	repo ports.StorageFileRepository
}

func NewStorageFileService(repo ports.StorageFileRepository) *StorageFileService {
	return &StorageFileService{repo: repo}
}

// UploadFile handles file upload with validation and unique filename generation
func (s *StorageFileService) UploadFile(ctx context.Context, file *entities.StorageFile) error {
	// Validate file size (1-10MB)
	const minSize = 1024 // 1KB minimum
	const maxSize = 10 * 1024 * 1024 // 10MB maximum

	if file.FileSize < minSize {
		return fmt.Errorf("file size too small: minimum is 1KB")
	}
	if file.FileSize > maxSize {
		return fmt.Errorf("file size too large: maximum is 10MB")
	}

	// Generate unique filename if not provided
	if file.Filename == "" {
		file.Filename = s.generateUniqueFilename(file.OriginalFilename)
	}

	// Ensure metadata is initialized
	if file.Metadata == nil {
		file.Metadata = make(map[string]interface{})
	}

	return s.repo.Create(ctx, file)
}

// GetFile retrieves a file by ID
func (s *StorageFileService) GetFile(ctx context.Context, id uuid.UUID) (*entities.StorageFile, error) {
	return s.repo.GetByID(ctx, id)
}

// GetFileByPath retrieves a file by folder and filename
func (s *StorageFileService) GetFileByPath(ctx context.Context, folder, filename string) (*entities.StorageFile, error) {
	return s.repo.GetByFilename(ctx, folder, filename)
}

// ListFilesByFolder lists all files in a specific folder
func (s *StorageFileService) ListFilesByFolder(ctx context.Context, folder string) ([]*entities.StorageFile, error) {
	return s.repo.ListByFolder(ctx, folder)
}

// ListFiles lists files with pagination
func (s *StorageFileService) ListFiles(ctx context.Context, limit, offset int) ([]*entities.StorageFile, error) {
	// Set default limit if not specified
	if limit <= 0 {
		limit = 50
	}

	// Cap maximum limit
	if limit > 100 {
		limit = 100
	}

	return s.repo.List(ctx, limit, offset)
}

// DeleteFile deletes a file by ID
func (s *StorageFileService) DeleteFile(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// UpdateFileMetadata updates the metadata of a file
func (s *StorageFileService) UpdateFileMetadata(ctx context.Context, id uuid.UUID, metadata map[string]interface{}) error {
	return s.repo.UpdateMetadata(ctx, id, metadata)
}

// generateUniqueFilename creates a unique filename while preserving the extension
func (s *StorageFileService) generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)

	// Generate random bytes for unique identifier
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)

	// Clean the original filename (remove extension and special chars)
	baseName := strings.TrimSuffix(originalFilename, ext)
	baseName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, baseName)

	// Truncate if too long
	if len(baseName) > 50 {
		baseName = baseName[:50]
	}

	return fmt.Sprintf("%s_%s%s", baseName, randomStr, ext)
}
