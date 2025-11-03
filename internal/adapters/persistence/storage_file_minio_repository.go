package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// StorageFileMinIORepository implements the StorageFileRepository interface using MinIO
type StorageFileMinIORepository struct {
	client     *minio.Client
	bucketName string
}

func NewStorageFileMinIORepository(client *minio.Client, bucketName string) *StorageFileMinIORepository {
	return &StorageFileMinIORepository{
		client:     client,
		bucketName: bucketName,
	}
}

// objectMetadata represents the metadata stored alongside the file in MinIO
type objectMetadata struct {
	ID               string                 `json:"id"`
	Filename         string                 `json:"filename"`
	Folder           string                 `json:"folder"`
	OriginalFilename string                 `json:"original_filename"`
	MimeType         string                 `json:"mime_type"`
	FileSize         int64                  `json:"file_size"`
	Metadata         map[string]interface{} `json:"metadata"`
	UploadedBy       string                 `json:"uploaded_by,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

func (r *StorageFileMinIORepository) Create(ctx context.Context, file *entities.StorageFile) error {
	// Generate UUID if not set
	if file.ID == uuid.Nil {
		file.ID = uuid.New()
	}

	// Create object path: folder/filename
	objectName := fmt.Sprintf("%s/%s", file.Folder, file.Filename)

	// Prepare user metadata for MinIO
	userMetadata := map[string]string{
		"x-amz-meta-id":                file.ID.String(),
		"x-amz-meta-folder":            file.Folder,
		"x-amz-meta-original-filename": file.OriginalFilename,
		"x-amz-meta-uploaded-by":       file.UploadedBy,
	}

	// Store custom metadata as JSON in a special metadata field
	if file.Metadata != nil {
		metadataJSON, err := json.Marshal(file.Metadata)
		if err == nil {
			userMetadata["x-amz-meta-custom-metadata"] = string(metadataJSON)
		}
	}

	// Upload the file
	_, err := r.client.PutObject(ctx, r.bucketName, objectName, bytes.NewReader(file.FileData), int64(len(file.FileData)), minio.PutObjectOptions{
		ContentType:  file.MimeType,
		UserMetadata: userMetadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	return nil
}

func (r *StorageFileMinIORepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.StorageFile, error) {
	// List all objects to find the one with matching ID
	objectCh := r.client.ListObjects(ctx, r.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// Get object info with metadata
		objInfo, err := r.client.StatObject(ctx, r.bucketName, object.Key, minio.StatObjectOptions{})
		if err != nil {
			continue
		}

		// Check if this is the object we're looking for
		if objInfo.UserMetadata["X-Amz-Meta-Id"] == id.String() {
			return r.getObjectByName(ctx, object.Key, objInfo)
		}
	}

	return nil, domainErrors.NewNotFoundError("StorageFile", id)
}

func (r *StorageFileMinIORepository) GetByFilename(ctx context.Context, folder, filename string) (*entities.StorageFile, error) {
	objectName := fmt.Sprintf("%s/%s", folder, filename)

	// Check if object exists
	objInfo, err := r.client.StatObject(ctx, r.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, domainErrors.NewNotFoundError("StorageFile", folder+"/"+filename)
		}
		return nil, err
	}

	return r.getObjectByName(ctx, objectName, objInfo)
}

func (r *StorageFileMinIORepository) ListByFolder(ctx context.Context, folder string) ([]*entities.StorageFile, error) {
	var files []*entities.StorageFile

	// List objects with prefix
	objectCh := r.client.ListObjects(ctx, r.bucketName, minio.ListObjectsOptions{
		Prefix:    folder + "/",
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// Get object info with metadata (but not the file data for listing)
		objInfo, err := r.client.StatObject(ctx, r.bucketName, object.Key, minio.StatObjectOptions{})
		if err != nil {
			continue
		}

		file := r.metadataToEntity(objInfo, nil)
		files = append(files, file)
	}

	return files, nil
}

func (r *StorageFileMinIORepository) List(ctx context.Context, limit, offset int) ([]*entities.StorageFile, error) {
	var files []*entities.StorageFile
	count := 0
	skipped := 0

	// List all objects
	objectCh := r.client.ListObjects(ctx, r.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// Skip until we reach offset
		if skipped < offset {
			skipped++
			continue
		}

		// Stop if we've reached the limit
		if limit > 0 && count >= limit {
			break
		}

		// Get object info with metadata
		objInfo, err := r.client.StatObject(ctx, r.bucketName, object.Key, minio.StatObjectOptions{})
		if err != nil {
			continue
		}

		file := r.metadataToEntity(objInfo, nil)
		files = append(files, file)
		count++
	}

	return files, nil
}

func (r *StorageFileMinIORepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Find the object by ID first
	objectCh := r.client.ListObjects(ctx, r.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}

		// Get object info with metadata
		objInfo, err := r.client.StatObject(ctx, r.bucketName, object.Key, minio.StatObjectOptions{})
		if err != nil {
			continue
		}

		// Check if this is the object we're looking for
		if objInfo.UserMetadata["X-Amz-Meta-Id"] == id.String() {
			err = r.client.RemoveObject(ctx, r.bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete file from MinIO: %w", err)
			}
			return nil
		}
	}

	return domainErrors.NewNotFoundError("StorageFile", id)
}

func (r *StorageFileMinIORepository) UpdateMetadata(ctx context.Context, id uuid.UUID, metadata map[string]interface{}) error {
	// Find the object by ID
	file, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	objectName := fmt.Sprintf("%s/%s", file.Folder, file.Filename)

	// Update metadata
	file.Metadata = metadata

	// Prepare updated user metadata
	userMetadata := map[string]string{
		"x-amz-meta-id":                file.ID.String(),
		"x-amz-meta-folder":            file.Folder,
		"x-amz-meta-original-filename": file.OriginalFilename,
		"x-amz-meta-uploaded-by":       file.UploadedBy,
	}

	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)
		if err == nil {
			userMetadata["x-amz-meta-custom-metadata"] = string(metadataJSON)
		}
	}

	// Copy object to itself with new metadata (MinIO way of updating metadata)
	src := minio.CopySrcOptions{
		Bucket: r.bucketName,
		Object: objectName,
	}

	dst := minio.CopyDestOptions{
		Bucket:          r.bucketName,
		Object:          objectName,
		UserMetadata:    userMetadata,
		ReplaceMetadata: true,
	}

	_, err = r.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// Helper methods

func (r *StorageFileMinIORepository) getObjectByName(ctx context.Context, objectName string, objInfo minio.ObjectInfo) (*entities.StorageFile, error) {
	// Download the file data
	object, err := r.client.GetObject(ctx, r.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	// Read file data
	fileData, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}

	return r.metadataToEntity(objInfo, fileData), nil
}

func (r *StorageFileMinIORepository) metadataToEntity(objInfo minio.ObjectInfo, fileData []byte) *entities.StorageFile {
	// MinIO returns user metadata in lowercase with underscores
	// Parse ID from metadata
	idStr := objInfo.UserMetadata["X-Amz-Meta-Id"]
	if idStr == "" {
		idStr = objInfo.UserMetadata["id"] // Try lowercase
	}
	id, _ := uuid.Parse(idStr)

	// Parse custom metadata
	var customMetadata map[string]interface{}
	metadataStr := objInfo.UserMetadata["X-Amz-Meta-Custom-Metadata"]
	if metadataStr == "" {
		metadataStr = objInfo.UserMetadata["custom-metadata"] // Try lowercase
	}
	if metadataStr != "" {
		json.Unmarshal([]byte(metadataStr), &customMetadata)
	}
	if customMetadata == nil {
		customMetadata = make(map[string]interface{})
	}

	// Extract folder and filename from key
	folder := objInfo.UserMetadata["X-Amz-Meta-Folder"]
	if folder == "" {
		folder = objInfo.UserMetadata["folder"] // Try lowercase
	}

	originalFilename := objInfo.UserMetadata["X-Amz-Meta-Original-Filename"]
	if originalFilename == "" {
		originalFilename = objInfo.UserMetadata["original-filename"] // Try lowercase
	}

	uploadedBy := objInfo.UserMetadata["X-Amz-Meta-Uploaded-By"]
	if uploadedBy == "" {
		uploadedBy = objInfo.UserMetadata["uploaded-by"] // Try lowercase
	}

	// Extract filename from full path (remove folder prefix)
	filename := objInfo.Key
	if folder != "" && len(objInfo.Key) > len(folder)+1 {
		filename = objInfo.Key[len(folder)+1:] // Remove "folder/" prefix
	}

	return &entities.StorageFile{
		ID:               id,
		Filename:         filename,
		Folder:           folder,
		OriginalFilename: originalFilename,
		MimeType:         objInfo.ContentType,
		FileSize:         objInfo.Size,
		FileData:         fileData,
		Metadata:         customMetadata,
		UploadedBy:       uploadedBy,
		CreatedAt:        objInfo.LastModified,
		UpdatedAt:        objInfo.LastModified,
	}
}
