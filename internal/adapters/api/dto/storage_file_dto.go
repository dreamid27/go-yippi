package dto

import "time"

// UploadFileRequest defines the request for uploading a file (JSON - legacy)
type UploadFileRequest struct {
	Folder     string `query:"folder" minLength:"1" doc:"Folder/service identifier" default:"general"`
	UploadedBy string `query:"uploaded_by" doc:"User identifier (optional)"`
	Body       struct {
		File      []byte                 `json:"file" doc:"Base64 encoded file data"`
		Filename  string                 `json:"filename" minLength:"1" doc:"Original filename"`
		MimeType  string                 `json:"mime_type" minLength:"1" doc:"File MIME type"`
		Metadata  map[string]interface{} `json:"metadata" doc:"Additional metadata (optional)"`
	}
}

// UploadFileMultipartRequest defines the request for uploading via multipart/form-data
type UploadFileMultipartRequest struct {
	Folder     string `query:"folder" doc:"Folder/service identifier" default:"general"`
	UploadedBy string `query:"uploaded_by" doc:"User identifier (optional)"`
	// Note: file and metadata come from multipart form, handled in handler
}

// StorageFileListItem represents a file item in lists (without file data)
type StorageFileListItem struct {
	ID               string                 `json:"id" doc:"File ID"`
	Filename         string                 `json:"filename" doc:"Generated unique filename"`
	Folder           string                 `json:"folder" doc:"Folder/service identifier"`
	OriginalFilename string                 `json:"original_filename" doc:"User's original filename"`
	MimeType         string                 `json:"mime_type" doc:"File MIME type"`
	FileSize         int64                  `json:"file_size" doc:"File size in bytes"`
	Metadata         map[string]interface{} `json:"metadata" doc:"Additional metadata"`
	UploadedBy       string                 `json:"uploaded_by,omitempty" doc:"User identifier"`
	CreatedAt        time.Time              `json:"created_at" doc:"Upload timestamp"`
	UpdatedAt        time.Time              `json:"updated_at" doc:"Last update timestamp"`
}

// FileUploadResponse defines the response after uploading a file
type FileUploadResponse struct {
	Body struct {
		ID               string                 `json:"id" doc:"File ID"`
		Filename         string                 `json:"filename" doc:"Generated unique filename"`
		Folder           string                 `json:"folder" doc:"Folder/service identifier"`
		OriginalFilename string                 `json:"original_filename" doc:"User's original filename"`
		MimeType         string                 `json:"mime_type" doc:"File MIME type"`
		FileSize         int64                  `json:"file_size" doc:"File size in bytes"`
		Metadata         map[string]interface{} `json:"metadata" doc:"Additional metadata"`
		UploadedBy       string                 `json:"uploaded_by,omitempty" doc:"User identifier"`
		CreatedAt        time.Time              `json:"created_at" doc:"Upload timestamp"`
		DownloadURL      string                 `json:"download_url" doc:"URL to download the file"`
	}
}

// GetFileRequest defines the request for getting a file by ID
type GetFileRequest struct {
	ID string `path:"id" doc:"File ID (UUID)"`
}

// GetFileByPathRequest defines the request for getting a file by folder and filename
type GetFileByPathRequest struct {
	Folder   string `path:"folder" doc:"Folder/service identifier"`
	Filename string `path:"filename" doc:"Filename"`
}

// FileDownloadResponse defines the response for downloading a file
type FileDownloadResponse struct {
	Body struct {
		ID               string                 `json:"id" doc:"File ID"`
		Filename         string                 `json:"filename" doc:"Generated unique filename"`
		OriginalFilename string                 `json:"original_filename" doc:"User's original filename"`
		MimeType         string                 `json:"mime_type" doc:"File MIME type"`
		FileSize         int64                  `json:"file_size" doc:"File size in bytes"`
		FileData         []byte                 `json:"file_data" doc:"Base64 encoded file content"`
		Metadata         map[string]interface{} `json:"metadata" doc:"Additional metadata"`
		UploadedBy       string                 `json:"uploaded_by,omitempty" doc:"User identifier"`
		CreatedAt        time.Time              `json:"created_at" doc:"Upload timestamp"`
	}
}

// ListFilesRequest defines the request for listing files
type ListFilesRequest struct {
	Folder string `query:"folder" doc:"Filter by folder (optional)"`
	Limit  int    `query:"limit" minimum:"1" maximum:"100" default:"50" doc:"Number of files to return"`
	Offset int    `query:"offset" minimum:"0" default:"0" doc:"Number of files to skip"`
}

// ListFilesResponse defines the response for listing files
type ListFilesResponse struct {
	Body struct {
		Files  []StorageFileListItem `json:"files" doc:"List of files"`
		Total  int                   `json:"total" doc:"Total number of files in result"`
		Limit  int                   `json:"limit" doc:"Limit used"`
		Offset int                   `json:"offset" doc:"Offset used"`
	}
}

// DeleteFileRequest defines the request for deleting a file
type DeleteFileRequest struct {
	ID string `path:"id" doc:"File ID (UUID)"`
}

// DeleteFileResponse defines the response for deleting a file
type DeleteFileResponse struct {
	Body struct {
		Success bool   `json:"success" doc:"Whether deletion was successful"`
		Message string `json:"message" doc:"Status message"`
	}
}

// UpdateMetadataRequest defines the request for updating file metadata
type UpdateMetadataRequest struct {
	ID   string `path:"id" doc:"File ID (UUID)"`
	Body struct {
		Metadata map[string]interface{} `json:"metadata" doc:"New metadata to set"`
	}
}

// UpdateMetadataResponse defines the response for updating metadata
type UpdateMetadataResponse struct {
	Body struct {
		Success bool   `json:"success" doc:"Whether update was successful"`
		Message string `json:"message" doc:"Status message"`
	}
}
