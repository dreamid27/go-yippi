package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStorageService is a mock implementation of StorageService
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) UploadFile(ctx context.Context, bucket, fileName string, reader io.Reader, size int64, contentType string) (*entities.FileMetadata, error) {
	args := m.Called(ctx, bucket, fileName, reader, size, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FileMetadata), args.Error(1)
}

func (m *MockStorageService) DeleteFile(ctx context.Context, bucket, fileName string) error {
	args := m.Called(ctx, bucket, fileName)
	return args.Error(0)
}

func (m *MockStorageService) GetFileURL(ctx context.Context, bucket, fileName string) (string, error) {
	args := m.Called(ctx, bucket, fileName)
	return args.String(0), args.Error(1)
}

// createMultipartFormData creates multipart form data for testing
func createMultipartFormData(file []byte, fileName, bucket, contentType string) []byte {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	part, _ := writer.CreateFormFile("file", fileName)
	part.Write(file)

	// Add file_name field if provided
	if fileName != "" {
		writer.WriteField("file_name", fileName)
	}

	// Add bucket field if provided
	if bucket != "" {
		writer.WriteField("bucket", bucket)
	}

	// Add content_type field if provided
	if contentType != "" {
		writer.WriteField("content_type", contentType)
	}

	writer.Close()
	return buf.Bytes()
}

func TestUploadFile_Success(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	fileContent := []byte("test file content")
	fileName := "test.txt"
	bucket := "test-bucket"
	contentType := "text/plain"

	input := &dto.UploadFileRequest{
		RawBody: createMultipartFormData(fileContent, fileName, bucket, contentType),
	}

	expectedMetadata := &entities.FileMetadata{
		ID:          "test-id-123",
		FileName:    fileName,
		Bucket:      bucket,
		Size:        int64(len(fileContent)),
		ContentType: contentType,
		URL:         "http://localhost:9000/test-bucket/test.txt",
		UploadedAt:  time.Now(),
	}

	mockService.On("UploadFile", ctx, bucket, fileName, mock.AnythingOfType("*bytes.Reader"), int64(len(fileContent)), contentType).Return(expectedMetadata, nil)

	// Act
	response, err := handler.UploadFile(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedMetadata.ID, response.Body.ID)
	assert.Equal(t, expectedMetadata.FileName, response.Body.FileName)
	assert.Equal(t, expectedMetadata.Bucket, response.Body.Bucket)
	assert.Equal(t, expectedMetadata.Size, response.Body.Size)
	assert.Equal(t, expectedMetadata.ContentType, response.Body.ContentType)
	assert.Equal(t, expectedMetadata.URL, response.Body.URL)
	mockService.AssertExpectations(t)
}

func TestUploadFile_AutoDetectContentType(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	// PNG file signature
	fileContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	fileName := "test.png"
	bucket := "test-bucket"

	input := &dto.UploadFileRequest{
		RawBody: createMultipartFormData(fileContent, fileName, bucket, ""),
	}

	expectedMetadata := &entities.FileMetadata{
		ID:          "test-id-123",
		FileName:    fileName,
		Bucket:      bucket,
		Size:        int64(len(fileContent)),
		ContentType: "image/png",
		URL:         "http://localhost:9000/test-bucket/test.png",
		UploadedAt:  time.Now(),
	}

	// Content type should be auto-detected as image/png
	mockService.On("UploadFile", ctx, bucket, fileName, mock.AnythingOfType("*bytes.Reader"), int64(len(fileContent)), "image/png").Return(expectedMetadata, nil)

	// Act
	response, err := handler.UploadFile(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "image/png", response.Body.ContentType)
	mockService.AssertExpectations(t)
}

func TestUploadFile_UseOriginalFilename(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	fileContent := []byte("test content")
	originalFileName := "original.txt"

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", originalFileName)
	part.Write(fileContent)
	writer.Close()

	input := &dto.UploadFileRequest{
		RawBody: buf.Bytes(),
	}

	expectedMetadata := &entities.FileMetadata{
		ID:          "test-id-123",
		FileName:    originalFileName,
		Bucket:      "",
		Size:        int64(len(fileContent)),
		ContentType: "text/plain; charset=utf-8",
		URL:         "http://localhost:9000/default/original.txt",
		UploadedAt:  time.Now(),
	}

	mockService.On("UploadFile", ctx, "", originalFileName, mock.AnythingOfType("*bytes.Reader"), int64(len(fileContent)), mock.Anything).Return(expectedMetadata, nil)

	// Act
	response, err := handler.UploadFile(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, originalFileName, response.Body.FileName)
	mockService.AssertExpectations(t)
}

func TestUploadFile_EmptyFile(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	input := &dto.UploadFileRequest{
		RawBody: createMultipartFormData([]byte{}, "test.txt", "", ""),
	}

	// Act
	response, err := handler.UploadFile(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	mockService.AssertNotCalled(t, "UploadFile")
}

func TestUploadFile_ServiceError(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	fileContent := []byte("test file content")
	fileName := "test.txt"

	input := &dto.UploadFileRequest{
		RawBody: createMultipartFormData(fileContent, fileName, "", ""),
	}

	mockService.On("UploadFile", ctx, "", fileName, mock.AnythingOfType("*bytes.Reader"), int64(len(fileContent)), mock.Anything).Return(nil, errors.New("storage error"))

	// Act
	response, err := handler.UploadFile(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 500, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

func TestUploadFile_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	fileContent := []byte("test file content")
	fileName := "test.txt"

	input := &dto.UploadFileRequest{
		RawBody: createMultipartFormData(fileContent, fileName, "", ""),
	}

	validationErr := domainErrors.NewValidationError("file_name", "invalid filename")
	mockService.On("UploadFile", ctx, "", fileName, mock.AnythingOfType("*bytes.Reader"), int64(len(fileContent)), mock.Anything).Return(nil, validationErr)

	// Act
	response, err := handler.UploadFile(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

func TestDeleteFile_Success(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	input := &dto.DeleteFileRequest{
		Bucket:   "test-bucket",
		FileName: "test.txt",
	}

	mockService.On("DeleteFile", ctx, "test-bucket", "test.txt").Return(nil)

	// Act
	response, err := handler.DeleteFile(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "File deleted successfully", response.Body.Message)
	mockService.AssertExpectations(t)
}

func TestDeleteFile_EmptyFileName(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	input := &dto.DeleteFileRequest{
		Bucket:   "test-bucket",
		FileName: "",
	}

	// Act
	response, err := handler.DeleteFile(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	mockService.AssertNotCalled(t, "DeleteFile")
}

func TestGetFileURL_Success(t *testing.T) {
	// Arrange
	mockService := new(MockStorageService)
	handler := NewFileHandler(mockService)
	ctx := context.Background()

	input := &dto.GetFileURLRequest{
		Bucket:   "test-bucket",
		FileName: "test.txt",
	}

	expectedURL := "http://localhost:9000/test-bucket/test.txt"
	mockService.On("GetFileURL", ctx, "test-bucket", "test.txt").Return(expectedURL, nil)

	// Act
	response, err := handler.GetFileURL(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedURL, response.Body.URL)
	mockService.AssertExpectations(t)
}

func TestMapToFileMetadataDTO(t *testing.T) {
	// Arrange
	uploadedAt := time.Now()
	metadata := &entities.FileMetadata{
		ID:          "test-id-123",
		FileName:    "test.txt",
		Bucket:      "test-bucket",
		Size:        1024,
		ContentType: "text/plain",
		URL:         "http://localhost:9000/test-bucket/test.txt",
		UploadedAt:  uploadedAt,
	}

	// Act
	result := mapToFileMetadataDTO(metadata)

	// Assert
	assert.Equal(t, metadata.ID, result.ID)
	assert.Equal(t, metadata.FileName, result.FileName)
	assert.Equal(t, metadata.Bucket, result.Bucket)
	assert.Equal(t, metadata.Size, result.Size)
	assert.Equal(t, metadata.ContentType, result.ContentType)
	assert.Equal(t, metadata.URL, result.URL)
	assert.Equal(t, metadata.UploadedAt, result.UploadedAt)
}

// TestExtractBoundary tests the boundary extraction helper
func TestExtractBoundary(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "Valid boundary",
			data:     "--WebKitFormBoundary123\r\nContent-Disposition: form-data",
			expected: "WebKitFormBoundary123",
		},
		{
			name:     "Empty data",
			data:     "",
			expected: "",
		},
		{
			name:     "No boundary",
			data:     "some random data",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBoundary([]byte(tt.data))
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a complete multipart request for debugging
func createCompleteMultipartRequest(file []byte, fileName, bucket string) string {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormFile("file", fileName)
	part.Write(file)

	if fileName != "" {
		writer.WriteField("file_name", fileName)
	}
	if bucket != "" {
		writer.WriteField("bucket", bucket)
	}

	writer.Close()

	return fmt.Sprintf("Content-Type: %s\r\n\r\n%s", writer.FormDataContentType(), buf.String())
}
