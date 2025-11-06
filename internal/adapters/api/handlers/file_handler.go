package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
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
}

// UploadFile handles file upload requests using multipart/form-data
func (h *FileHandler) UploadFile(ctx context.Context, input *dto.UploadFileRequest) (*dto.FileMetadataResponse, error) {
	// Parse multipart form data from raw body
	reader := bytes.NewReader(input.RawBody)

	// Extract boundary from content-type header (Huma should handle this)
	// For now, we'll parse the multipart form manually
	multipartReader := multipart.NewReader(reader, extractBoundary(input.RawBody))

	var fileData []byte
	var fileName string
	var bucket string
	var contentType string

	// Parse all form fields
	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, huma.Error400BadRequest("failed to parse multipart form: " + err.Error())
		}

		fieldName := part.FormName()

		switch fieldName {
		case "file":
			// Read file content
			fileData, err = io.ReadAll(part)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read file content: " + err.Error())
			}
			// Get original filename if not provided separately
			if fileName == "" && part.FileName() != "" {
				fileName = part.FileName()
			}
			// Auto-detect content type from file if not provided
			if contentType == "" {
				contentType = http.DetectContentType(fileData)
			}

		case "file_name":
			// Read custom filename
			data, err := io.ReadAll(part)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read file_name field: " + err.Error())
			}
			fileName = string(data)

		case "bucket":
			// Read bucket name
			data, err := io.ReadAll(part)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read bucket field: " + err.Error())
			}
			bucket = string(data)

		case "content_type":
			// Read custom content type
			data, err := io.ReadAll(part)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read content_type field: " + err.Error())
			}
			contentType = string(data)
		}

		part.Close()
	}

	// Validate required fields
	if len(fileData) == 0 {
		return nil, huma.Error400BadRequest("file is required")
	}

	if fileName == "" {
		return nil, huma.Error400BadRequest("file_name is required (either as form field or from uploaded file)")
	}

	// Default content type if still not set
	if contentType == "" {
		contentType = "application/octet-stream"
	}

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

// extractBoundary extracts the boundary string from multipart form data
func extractBoundary(data []byte) string {
	// Simple boundary extraction - look for the first boundary in the data
	// Format: ------WebKitFormBoundary...
	dataStr := string(data)
	if len(dataStr) < 2 {
		return ""
	}

	// Find first boundary (starts with --)
	start := bytes.Index(data, []byte("--"))
	if start == -1 {
		return ""
	}

	// Find end of boundary (CR LF)
	end := bytes.Index(data[start:], []byte("\r\n"))
	if end == -1 {
		end = bytes.Index(data[start:], []byte("\n"))
	}

	if end == -1 {
		return ""
	}

	// Extract boundary without the leading --
	boundary := string(data[start+2 : start+end])
	return boundary
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
