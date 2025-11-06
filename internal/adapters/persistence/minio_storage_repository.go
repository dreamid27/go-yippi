package persistence

import (
	"context"
	"fmt"
	"io"
	"time"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// MinIOStorageRepository implements StorageRepository using MinIO
type MinIOStorageRepository struct {
	client   *minio.Client
	endpoint string
	useSSL   bool
}

// NewMinIOStorageRepository creates a new MinIO storage repository
func NewMinIOStorageRepository(client *minio.Client, endpoint string, useSSL bool) *MinIOStorageRepository {
	return &MinIOStorageRepository{
		client:   client,
		endpoint: endpoint,
		useSSL:   useSSL,
	}
}

// Store uploads a file to MinIO and returns metadata
func (r *MinIOStorageRepository) Store(ctx context.Context, bucket, fileName string, reader io.Reader, size int64, contentType string) (*entities.FileMetadata, error) {
	// Upload file to MinIO
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	info, err := r.client.PutObject(ctx, bucket, fileName, reader, size, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	// Generate file URL
	url, err := r.GetURL(ctx, bucket, fileName)
	if err != nil {
		return nil, err
	}

	// Create metadata
	metadata := &entities.FileMetadata{
		ID:          uuid.New().String(),
		FileName:    fileName,
		Bucket:      bucket,
		Size:        info.Size,
		ContentType: contentType,
		URL:         url,
		UploadedAt:  time.Now(),
	}

	return metadata, nil
}

// Remove deletes a file from MinIO
func (r *MinIOStorageRepository) Remove(ctx context.Context, bucket, fileName string) error {
	err := r.client.RemoveObject(ctx, bucket, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	return nil
}

// GetURL generates a relative URL for the file that will be proxied through the API
func (r *MinIOStorageRepository) GetURL(ctx context.Context, bucket, fileName string) (string, error) {
	// Return relative URL that points to our download endpoint
	// This ensures all requests go through our service instead of directly to MinIO
	url := fmt.Sprintf("/files/download?bucket=%s&file_name=%s", bucket, fileName)
	return url, nil
}

// GetFile retrieves a file from MinIO and returns its content, size, and content type
func (r *MinIOStorageRepository) GetFile(ctx context.Context, bucket, fileName string) (io.ReadCloser, int64, string, error) {
	// Get file object from MinIO
	object, err := r.client.GetObject(ctx, bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to get file from MinIO: %w", err)
	}

	// Get file stats to retrieve size and content type
	stat, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, 0, "", fmt.Errorf("failed to get file stats from MinIO: %w", err)
	}

	return object, stat.Size, stat.ContentType, nil
}

// EnsureBucket creates a bucket if it doesn't exist
func (r *MinIOStorageRepository) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := r.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = r.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}
