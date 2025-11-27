package ports

import (
	"context"
	"io"

	"example.com/go-yippi/internal/domain/entities"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	UploadFile(ctx context.Context, bucket, fileName string, reader io.Reader, size int64, contentType string) (*entities.FileMetadata, error)
	DeleteFile(ctx context.Context, bucket, fileName string) error
	GetFileURL(ctx context.Context, bucket, fileName string) (string, error)
	DownloadFile(ctx context.Context, bucket, fileName string) (io.ReadCloser, int64, string, error)
}
