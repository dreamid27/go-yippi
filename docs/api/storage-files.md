# Storage File Service

A complete file storage service implementation following hexagonal architecture principles. This service allows you to upload files (1KB-10MB) directly to PostgreSQL, download them, organize them by folders, and manage metadata.

## Features

- ✅ Upload files (1KB-10MB) directly to PostgreSQL
- ✅ Download files by ID or URL path (folder/filename)
- ✅ Organize files by folder/service identifier
- ✅ Delete files
- ✅ Store and update metadata with files
- ✅ List files with pagination and folder filtering
- ✅ Automatic unique filename generation
- ✅ Full OpenAPI documentation

## Database Schema

The service uses a `storage_files` table with the following structure:

```sql
Table "public.storage_files"
      Column       |           Type           | Nullable | Default
-------------------+--------------------------+----------+---------
 id                | uuid                     | not null |
 filename          | character varying        | not null |
 folder            | character varying        | not null |
 original_filename | character varying        | not null |
 mime_type         | character varying        | not null |
 file_size         | bigint                   | not null |
 file_data         | bytea                    | not null |
 metadata          | jsonb                    |          |
 uploaded_by       | character varying        |          |
 created_at        | timestamp with time zone | not null |
 updated_at        | timestamp with time zone | not null |

Indexes:
    "storage_files_pkey" PRIMARY KEY, btree (id)
    "storagefile_created_at" btree (created_at)
    "storagefile_filename" btree (filename)
    "storagefile_folder" btree (folder)
```

## API Endpoints

### 1. Upload File

**POST** `/files?folder={folder}&uploaded_by={user}`

Upload a file to storage.

**Query Parameters:**
- `folder` (optional, default: "general") - Folder/service identifier
- `uploaded_by` (optional) - User identifier

**Request Body:**
```json
{
  "file": "base64_encoded_file_data",
  "filename": "document.pdf",
  "mime_type": "application/pdf",
  "metadata": {
    "description": "Important document",
    "version": "1.0"
  }
}
```

**Response:**
```json
{
  "id": "956d001c-9e38-4366-8cd7-abf241d498de",
  "filename": "document_9e49054f0805972373732b89e0264ed0.pdf",
  "folder": "documents",
  "original_filename": "document.pdf",
  "mime_type": "application/pdf",
  "file_size": 2048,
  "metadata": {
    "description": "Important document",
    "version": "1.0"
  },
  "uploaded_by": "user123",
  "created_at": "2025-11-02T04:25:37.091421Z",
  "download_url": "/files/956d001c-9e38-4366-8cd7-abf241d498de"
}
```

**Validation:**
- File size must be between 1KB and 10MB
- Filename, MIME type, and file data are required

### 2. Download File by ID

**GET** `/files/{id}`

Download a file using its UUID.

**Response:**
```json
{
  "id": "956d001c-9e38-4366-8cd7-abf241d498de",
  "filename": "document_9e49054f0805972373732b89e0264ed0.pdf",
  "original_filename": "document.pdf",
  "mime_type": "application/pdf",
  "file_size": 2048,
  "file_data": "base64_encoded_file_content",
  "metadata": {
    "description": "Important document"
  },
  "uploaded_by": "user123",
  "created_at": "2025-11-02T04:25:37.091421Z"
}
```

### 3. Download File by Path

**GET** `/files/{folder}/{filename}`

Download a file using folder and filename.

**Example:** `/files/documents/document_9e49054f0805972373732b89e0264ed0.pdf`

### 4. List Files

**GET** `/files?folder={folder}&limit={limit}&offset={offset}`

List files with optional filtering and pagination.

**Query Parameters:**
- `folder` (optional) - Filter by folder
- `limit` (optional, default: 50, max: 100) - Number of files to return
- `offset` (optional, default: 0) - Number of files to skip

**Response:**
```json
{
  "files": [
    {
      "id": "956d001c-9e38-4366-8cd7-abf241d498de",
      "filename": "document_9e49054f0805972373732b89e0264ed0.pdf",
      "folder": "documents",
      "original_filename": "document.pdf",
      "mime_type": "application/pdf",
      "file_size": 2048,
      "metadata": {},
      "uploaded_by": "user123",
      "created_at": "2025-11-02T04:25:37.091421Z",
      "updated_at": "2025-11-02T04:25:37.091421Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

### 5. Update File Metadata

**PATCH** `/files/{id}/metadata`

Update the metadata of a file.

**Request Body:**
```json
{
  "metadata": {
    "description": "Updated description",
    "version": "2.0",
    "tags": ["important", "reviewed"]
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Metadata updated successfully"
}
```

### 6. Delete File

**DELETE** `/files/{id}`

Delete a file from storage.

**Response:**
```json
{
  "success": true,
  "message": "File deleted successfully"
}
```

## Architecture

The service follows hexagonal architecture with clear layer separation:

### Domain Layer (`internal/domain/`)
- **Entity:** `entities/storage_file.go` - Pure domain entity
- **Port:** `ports/repository.go` - StorageFileRepository interface

### Application Layer (`internal/application/services/`)
- **Service:** `storage_file_service.go` - Business logic
  - File size validation (1KB-10MB)
  - Unique filename generation
  - Metadata management

### Adapters Layer (`internal/adapters/`)
- **Persistence:** `persistence/storage_file_repository.go` - Ent repository implementation
- **API:** `api/handlers/storage_file_handler.go` - HTTP handlers
- **DTOs:** `api/dto/storage_file_dto.go` - Request/response types

### Infrastructure Layer
- **Schema:** `persistence/db/schema/storage_file.go` - Ent schema definition

## Usage Examples

### Upload a text file

```bash
# Create base64 encoded content
echo "Hello World! This is a test file." | base64 > /tmp/file_data.txt

# Upload the file
curl -X POST "http://localhost:8080/files?folder=demo&uploaded_by=user123" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"$(cat /tmp/file_data.txt)\",
    \"filename\": \"hello.txt\",
    \"mime_type\": \"text/plain\",
    \"metadata\": {\"type\": \"demo\"}
  }"
```

### Download a file

```bash
# By ID
curl -X GET "http://localhost:8080/files/{file-id}"

# By path
curl -X GET "http://localhost:8080/files/demo/hello_abc123.txt"
```

### List files in a folder

```bash
curl -X GET "http://localhost:8080/files?folder=demo&limit=10"
```

### Update metadata

```bash
curl -X PATCH "http://localhost:8080/files/{file-id}/metadata" \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "tags": ["important", "reviewed"],
      "status": "approved"
    }
  }'
```

### Delete a file

```bash
curl -X DELETE "http://localhost:8080/files/{file-id}"
```

## Testing

A test script is provided at `test_storage_api.sh` to demonstrate all endpoints:

```bash
chmod +x test_storage_api.sh
./test_storage_api.sh
```

## File Size Limits

- **Minimum:** 1KB
- **Maximum:** 10MB

Files outside this range will be rejected with a 400 Bad Request error.

## Unique Filename Generation

The service automatically generates unique filenames while preserving the original extension:

**Original:** `document.pdf`
**Generated:** `document_9e49054f0805972373732b89e0264ed0.pdf`

This ensures:
- No filename collisions
- Original extension is preserved
- Files remain identifiable
- Safe characters only (alphanumeric, hyphens, underscores)

## OpenAPI Documentation

When the server is running, full OpenAPI documentation is available at:
```
http://localhost:8080/docs
```

The documentation includes:
- All endpoints with detailed descriptions
- Request/response schemas
- Example values
- Try-it-out functionality

## Error Handling

The service uses domain errors for proper HTTP status mapping:

- **400 Bad Request** - Invalid file size, malformed UUID, validation errors
- **404 Not Found** - File not found
- **500 Internal Server Error** - Database or server errors

All errors follow Huma's error response format:
```json
{
  "title": "Bad Request",
  "status": 400,
  "detail": "file size too large: maximum is 10MB"
}
```
