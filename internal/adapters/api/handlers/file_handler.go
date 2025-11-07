package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/danielgtaylor/huma/v2"
)

// FileHandler handles HTTP requests for file storage operations
type FileHandler struct {
	service ports.StorageService
}

// NewFileHandler creates a new file handler
func NewFileHandler(service ports.StorageService) *FileHandler {
	return &FileHandler{service: service}
}

// RegisterRoutes registers all file storage routes with Huma
func (h *FileHandler) RegisterRoutes(api huma.API) {
	// Upload file
	huma.Register(api, huma.Operation{
		OperationID: "upload-file",
		Method:      http.MethodPost,
		Path:        "/files/upload",
		Summary:     "Upload a file to storage",
		Description: "Uploads a file to MinIO storage with custom filename and bucket selection",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, h.UploadFile)

	// Delete file
	huma.Register(api, huma.Operation{
		OperationID: "delete-file",
		Method:      http.MethodDelete,
		Path:        "/files",
		Summary:     "Delete a file from storage",
		Description: "Deletes a file from MinIO storage by filename and bucket",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, h.DeleteFile)

	// Get file URL
	huma.Register(api, huma.Operation{
		OperationID: "get-file-url",
		Method:      http.MethodGet,
		Path:        "/files/url",
		Summary:     "Get file URL",
		Description: "Retrieves the public URL for a file in storage",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, h.GetFileURL)

	// Download file
	huma.Register(api, huma.Operation{
		OperationID: "download-file",
		Method:      http.MethodGet,
		Path:        "/files/download",
		Summary:     "Download a file",
		Description: "Downloads a file from storage and streams it to the client",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
	}, h.DownloadFile)
}

// UploadFile handles file upload requests using multipart/form-data
func (h *FileHandler) UploadFile(ctx context.Context, input *dto.UploadFileRequest) (*dto.FileMetadataResponse, error) {
	// Get form data from multipart request
	formData := input.RawBody.Data()

	// Check if file was provided
	if !formData.File.IsSet {
		return nil, huma.Error400BadRequest("file is required")
	}

	// Read file content
	fileData, err := io.ReadAll(formData.File.File)
	if err != nil {
		return nil, huma.Error400BadRequest("failed to read file content: " + err.Error())
	}

	// Determine filename: use custom file_name if provided, otherwise use uploaded filename
	fileName := formData.FileName
	if fileName == "" {
		fileName = formData.File.Filename
	}

	if fileName == "" {
		return nil, huma.Error400BadRequest("file_name is required (either as form field or from uploaded file)")
	}

	// Determine content type: use custom content_type if provided, otherwise auto-detect
	contentType := formData.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(fileData)
	}

	// Default content type if still not detected
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Use provided bucket or empty string for default
	bucket := formData.Bucket

	// Create reader from file bytes
	fileReader := bytes.NewReader(fileData)
	size := int64(len(fileData))

	// Upload file
	metadata, err := h.service.UploadFile(ctx, bucket, fileName, fileReader, size, contentType)
	if err != nil {
		// Handle domain errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}

		return nil, huma.Error500InternalServerError("failed to upload file", err)
	}

	// Map to DTO
	response := &dto.FileMetadataResponse{
		Body: dto.FileMetadataDTO{
			ID:          metadata.ID,
			FileName:    metadata.FileName,
			Bucket:      metadata.Bucket,
			Size:        metadata.Size,
			ContentType: metadata.ContentType,
			URL:         metadata.URL,
			UploadedAt:  metadata.UploadedAt,
		},
	}

	return response, nil
}

// DeleteFile handles file deletion requests
func (h *FileHandler) DeleteFile(ctx context.Context, input *dto.DeleteFileRequest) (*dto.DeleteFileResponse, error) {
	// Validate filename
	if input.FileName == "" {
		return nil, huma.Error400BadRequest("file_name is required")
	}

	// Delete file
	err := h.service.DeleteFile(ctx, input.Bucket, input.FileName)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to delete file", err)
	}

	response := &dto.DeleteFileResponse{}
	response.Body.Message = "File deleted successfully"

	return response, nil
}

// GetFileURL handles file URL retrieval requests
func (h *FileHandler) GetFileURL(ctx context.Context, input *dto.GetFileURLRequest) (*dto.FileURLResponse, error) {
	// Validate filename
	if input.FileName == "" {
		return nil, huma.Error400BadRequest("file_name is required")
	}

	// Get file URL
	url, err := h.service.GetFileURL(ctx, input.Bucket, input.FileName)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get file URL", err)
	}

	response := &dto.FileURLResponse{}
	response.Body.URL = url

	return response, nil
}

// DownloadFile handles file download requests and streams the file to the client
func (h *FileHandler) DownloadFile(ctx context.Context, input *dto.DownloadFileRequest) (*huma.StreamResponse, error) {
	// Validate filename
	if input.FileName == "" {
		return nil, huma.Error400BadRequest("file_name is required")
	}

	// Download file
	reader, size, contentType, err := h.service.DownloadFile(ctx, input.Bucket, input.FileName)
	if err != nil {
		// Handle domain errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}

		return nil, huma.Error500InternalServerError("failed to download file", err)
	}

	// Return stream response
	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			defer reader.Close()

			// Set content type and content length headers
			ctx.SetHeader("Content-Type", contentType)
			ctx.SetHeader("Content-Length", fmt.Sprintf("%d", size))
			ctx.SetHeader("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", input.FileName))

			// Stream file content to response
			io.Copy(ctx.BodyWriter(), reader)
		},
	}, nil
}

// mapToFileMetadataDTO maps domain entity to DTO
func mapToFileMetadataDTO(metadata *entities.FileMetadata) dto.FileMetadataDTO {
	return dto.FileMetadataDTO{
		ID:          metadata.ID,
		FileName:    metadata.FileName,
		Bucket:      metadata.Bucket,
		Size:        metadata.Size,
		ContentType: metadata.ContentType,
		URL:         metadata.URL,
		UploadedAt:  metadata.UploadedAt,
	}
}
