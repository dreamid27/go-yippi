package entities

import "time"

// FileMetadata represents metadata for a stored file
type FileMetadata struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	Bucket      string    `json:"bucket"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	URL         string    `json:"url"`
	UploadedAt  time.Time `json:"uploaded_at"`
}
