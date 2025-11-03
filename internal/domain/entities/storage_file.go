package entities

import (
	"time"

	"github.com/google/uuid"
)

// StorageFile represents a file stored in the system
type StorageFile struct {
	ID               uuid.UUID
	Filename         string    // Generated unique filename
	Folder           string    // Service identifier/folder
	OriginalFilename string    // User's original filename
	MimeType         string    // File MIME type
	FileSize         int64     // Size in bytes
	FileData         []byte    // Binary file content
	Metadata         map[string]interface{} // Additional metadata
	UploadedBy       string    // User identifier
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
