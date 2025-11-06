package dto

import (
	"time"
)

// UploadFileRequest represents the request to upload a file using multipart/form-data
type UploadFileRequest struct {
	RawBody []byte `contentType:"multipart/form-data"`
}

// FileMetadataResponse represents the response containing file metadata
type FileMetadataResponse struct {
	Body FileMetadataDTO
}

// FileMetadataDTO represents file metadata in the response
type FileMetadataDTO struct {
	ID          string    `json:"id" doc:"Unique file identifier"`
	FileName    string    `json:"file_name" doc:"Name of the uploaded file"`
	Bucket      string    `json:"bucket" doc:"Bucket where the file is stored"`
	Size        int64     `json:"size" doc:"File size in bytes"`
	ContentType string    `json:"content_type" doc:"MIME type of the file"`
	URL         string    `json:"url" doc:"Public URL to access the file"`
	UploadedAt  time.Time `json:"uploaded_at" doc:"Timestamp when the file was uploaded"`
}

// DeleteFileRequest represents the request to delete a file
type DeleteFileRequest struct {
	Bucket   string `query:"bucket" doc:"Bucket name (optional, uses default if not specified)"`
	FileName string `query:"file_name" required:"true" doc:"Name of the file to delete"`
}

// DeleteFileResponse represents the response after deleting a file
type DeleteFileResponse struct {
	Body struct {
		Message string `json:"message" doc:"Confirmation message"`
	}
}

// GetFileURLRequest represents the request to get a file URL
type GetFileURLRequest struct {
	Bucket   string `query:"bucket" doc:"Bucket name (optional, uses default if not specified)"`
	FileName string `query:"file_name" required:"true" doc:"Name of the file"`
}

// FileURLResponse represents the response containing a file URL
type FileURLResponse struct {
	Body struct {
		URL string `json:"url" doc:"Public URL to access the file"`
	}
}

// DownloadFileRequest represents the request to download a file
type DownloadFileRequest struct {
	Bucket   string `query:"bucket" doc:"Bucket name (optional, uses default if not specified)"`
	FileName string `query:"file_name" required:"true" doc:"Name of the file to download"`
}
