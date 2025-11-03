package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/application/services"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// StorageFileHandler handles HTTP requests for file storage
type StorageFileHandler struct {
	service *services.StorageFileService
}

func NewStorageFileHandler(service *services.StorageFileService) *StorageFileHandler {
	return &StorageFileHandler{service: service}
}

// RegisterRoutes registers all storage file routes with Huma
func (h *StorageFileHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "upload-file-json",
		Method:      http.MethodPost,
		Path:        "/files/upload-json",
		Summary:     "Upload a file (JSON/Base64)",
		Description: "Upload a file using JSON with base64 encoded data (legacy method, 1-10MB)",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, h.UploadFile)

	huma.Register(api, huma.Operation{
		OperationID: "get-file-by-id",
		Method:      http.MethodGet,
		Path:        "/files/{id}",
		Summary:     "Download a file by ID",
		Description: "Download a file using its UUID",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
	}, h.GetFileByID)

	huma.Register(api, huma.Operation{
		OperationID: "get-file-by-path",
		Method:      http.MethodGet,
		Path:        "/files/{folder}/{filename}",
		Summary:     "Download a file by path",
		Description: "Download a file using folder and filename",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetFileByPath)

	huma.Register(api, huma.Operation{
		OperationID: "list-files",
		Method:      http.MethodGet,
		Path:        "/files",
		Summary:     "List files",
		Description: "List all files with optional folder filter and pagination",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.ListFiles)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-file",
		Method:        http.MethodDelete,
		Path:          "/files/{id}",
		Summary:       "Delete a file",
		Description:   "Delete a file from storage",
		Tags:          []string{"Files"},
		DefaultStatus: http.StatusOK,
		Errors:        []int{http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
	}, h.DeleteFile)

	huma.Register(api, huma.Operation{
		OperationID: "update-file-metadata",
		Method:      http.MethodPatch,
		Path:        "/files/{id}/metadata",
		Summary:     "Update file metadata",
		Description: "Update the metadata of a stored file",
		Tags:        []string{"Files"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
	}, h.UpdateMetadata)
}

// RegisterFiberRoutes registers routes that need direct Fiber access (like multipart uploads)
func (h *StorageFileHandler) RegisterFiberRoutes(app *fiber.App) {
	app.Post("/files", h.UploadFileMultipart)
}

func (h *StorageFileHandler) UploadFile(ctx context.Context, input *dto.UploadFileRequest) (*dto.FileUploadResponse, error) {
	// Set default folder if empty
	folder := input.Folder
	if folder == "" {
		folder = "general"
	}

	file := &entities.StorageFile{
		Folder:           folder,
		OriginalFilename: input.Body.Filename,
		MimeType:         input.Body.MimeType,
		FileSize:         int64(len(input.Body.File)),
		FileData:         input.Body.File,
		Metadata:         input.Body.Metadata,
		UploadedBy:       input.UploadedBy,
	}

	err := h.service.UploadFile(ctx, file)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}

	resp := &dto.FileUploadResponse{}
	resp.Body.ID = file.ID.String()
	resp.Body.Filename = file.Filename
	resp.Body.Folder = file.Folder
	resp.Body.OriginalFilename = file.OriginalFilename
	resp.Body.MimeType = file.MimeType
	resp.Body.FileSize = file.FileSize
	resp.Body.Metadata = file.Metadata
	resp.Body.UploadedBy = file.UploadedBy
	resp.Body.CreatedAt = file.CreatedAt
	resp.Body.DownloadURL = fmt.Sprintf("/files/%s", file.ID.String())

	return resp, nil
}

func (h *StorageFileHandler) GetFileByID(ctx context.Context, input *dto.GetFileRequest) (*dto.FileDownloadResponse, error) {
	// Parse UUID
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid file ID format")
	}

	file, err := h.service.GetFile(ctx, id)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("File not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get file", err)
	}

	resp := &dto.FileDownloadResponse{}
	resp.Body.ID = file.ID.String()
	resp.Body.Filename = file.Filename
	resp.Body.OriginalFilename = file.OriginalFilename
	resp.Body.MimeType = file.MimeType
	resp.Body.FileSize = file.FileSize
	resp.Body.FileData = file.FileData
	resp.Body.Metadata = file.Metadata
	resp.Body.UploadedBy = file.UploadedBy
	resp.Body.CreatedAt = file.CreatedAt

	return resp, nil
}

func (h *StorageFileHandler) GetFileByPath(ctx context.Context, input *dto.GetFileByPathRequest) (*dto.FileDownloadResponse, error) {
	file, err := h.service.GetFileByPath(ctx, input.Folder, input.Filename)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("File not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get file", err)
	}

	resp := &dto.FileDownloadResponse{}
	resp.Body.ID = file.ID.String()
	resp.Body.Filename = file.Filename
	resp.Body.OriginalFilename = file.OriginalFilename
	resp.Body.MimeType = file.MimeType
	resp.Body.FileSize = file.FileSize
	resp.Body.FileData = file.FileData
	resp.Body.Metadata = file.Metadata
	resp.Body.UploadedBy = file.UploadedBy
	resp.Body.CreatedAt = file.CreatedAt

	return resp, nil
}

func (h *StorageFileHandler) ListFiles(ctx context.Context, input *dto.ListFilesRequest) (*dto.ListFilesResponse, error) {
	var files []*entities.StorageFile
	var err error

	if input.Folder != "" {
		// List by folder
		files, err = h.service.ListFilesByFolder(ctx, input.Folder)
	} else {
		// List all with pagination
		files, err = h.service.ListFiles(ctx, input.Limit, input.Offset)
	}

	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list files", err)
	}

	resp := &dto.ListFilesResponse{}
	resp.Body.Files = make([]dto.StorageFileListItem, len(files))

	for i, file := range files {
		resp.Body.Files[i] = dto.StorageFileListItem{
			ID:               file.ID.String(),
			Filename:         file.Filename,
			Folder:           file.Folder,
			OriginalFilename: file.OriginalFilename,
			MimeType:         file.MimeType,
			FileSize:         file.FileSize,
			Metadata:         file.Metadata,
			UploadedBy:       file.UploadedBy,
			CreatedAt:        file.CreatedAt,
			UpdatedAt:        file.UpdatedAt,
		}
	}

	resp.Body.Total = len(files)
	resp.Body.Limit = input.Limit
	resp.Body.Offset = input.Offset

	return resp, nil
}

func (h *StorageFileHandler) DeleteFile(ctx context.Context, input *dto.DeleteFileRequest) (*dto.DeleteFileResponse, error) {
	// Parse UUID
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid file ID format")
	}

	err = h.service.DeleteFile(ctx, id)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("File not found")
		}
		return nil, huma.Error500InternalServerError("Failed to delete file", err)
	}

	resp := &dto.DeleteFileResponse{}
	resp.Body.Success = true
	resp.Body.Message = "File deleted successfully"

	return resp, nil
}

func (h *StorageFileHandler) UpdateMetadata(ctx context.Context, input *dto.UpdateMetadataRequest) (*dto.UpdateMetadataResponse, error) {
	// Parse UUID
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid file ID format")
	}

	err = h.service.UpdateFileMetadata(ctx, id, input.Body.Metadata)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("File not found")
		}
		return nil, huma.Error500InternalServerError("Failed to update metadata", err)
	}

	resp := &dto.UpdateMetadataResponse{}
	resp.Body.Success = true
	resp.Body.Message = "Metadata updated successfully"

	return resp, nil
}

// UploadFileMultipart handles multipart/form-data file uploads
func (h *StorageFileHandler) UploadFileMultipart(c *fiber.Ctx) error {
	// Get query parameters
	folder := c.Query("folder", "general")
	uploadedBy := c.Query("uploaded_by", "")

	// Get the file from multipart form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file",
		})
	}

	// Get optional metadata from form field
	var metadata map[string]interface{}
	metadataStr := c.FormValue("metadata")
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid metadata JSON",
			})
		}
	}

	// Create storage file entity
	storageFile := &entities.StorageFile{
		Folder:           folder,
		OriginalFilename: fileHeader.Filename,
		MimeType:         fileHeader.Header.Get("Content-Type"),
		FileSize:         int64(len(fileData)),
		FileData:         fileData,
		Metadata:         metadata,
		UploadedBy:       uploadedBy,
	}

	// If MIME type is empty, try to detect from filename
	if storageFile.MimeType == "" {
		storageFile.MimeType = "application/octet-stream"
	}

	// Upload file using service
	err = h.service.UploadFile(c.Context(), storageFile)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":                storageFile.ID.String(),
		"filename":          storageFile.Filename,
		"folder":            storageFile.Folder,
		"original_filename": storageFile.OriginalFilename,
		"mime_type":         storageFile.MimeType,
		"file_size":         storageFile.FileSize,
		"metadata":          storageFile.Metadata,
		"uploaded_by":       storageFile.UploadedBy,
		"created_at":        storageFile.CreatedAt,
		"download_url":      fmt.Sprintf("/files/%s", storageFile.ID.String()),
	})
}
