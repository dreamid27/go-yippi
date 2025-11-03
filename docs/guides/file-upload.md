# File Upload Guide

The Storage File Service supports **two methods** for file uploads:

1. **Multipart/Form-Data** (Recommended) - Standard HTTP file upload
2. **JSON/Base64** (Legacy) - Base64 encoded files in JSON

## Method 1: Multipart/Form-Data Upload (Recommended)

This is the **standard and recommended** way to upload files via HTTP.

### Endpoint

```
POST /files?folder={folder}&uploaded_by={user}
Content-Type: multipart/form-data
```

### Parameters

**Query Parameters:**
- `folder` (optional, default: "general") - Folder/service identifier
- `uploaded_by` (optional) - User identifier

**Form Fields:**
- `file` (required) - The file to upload
- `metadata` (optional) - JSON string with additional metadata

### Examples

#### Upload a Text File

```bash
curl -X POST "http://localhost:8080/files?folder=documents&uploaded_by=john" \
  -F "file=@/path/to/document.pdf" \
  -F 'metadata={"category": "invoice", "year": 2024}'
```

#### Upload an Image

```bash
curl -X POST "http://localhost:8080/files?folder=avatars&uploaded_by=user123" \
  -F "file=@/path/to/photo.jpg"
```

#### Upload with cURL (Detailed)

```bash
curl -X POST "http://localhost:8080/files?folder=attachments" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.pdf" \
  -F 'metadata={"type": "contract", "status": "draft"}'
```

### Response

```json
{
  "id": "4b21c755-3c9b-4e84-9e15-67a4fa88bab8",
  "filename": "document_65833dffc8a74268bffa6dfb92655d83.pdf",
  "folder": "documents",
  "original_filename": "document.pdf",
  "mime_type": "application/pdf",
  "file_size": 45678,
  "metadata": {
    "category": "invoice",
    "year": 2024
  },
  "uploaded_by": "john",
  "created_at": "2025-11-02T20:21:44Z",
  "download_url": "/files/4b21c755-3c9b-4e84-9e15-67a4fa88bab8"
}
```

### JavaScript/TypeScript Example

```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);
formData.append('metadata', JSON.stringify({
  category: 'document',
  confidential: true
}));

const response = await fetch('http://localhost:8080/files?folder=uploads&uploaded_by=user123', {
  method: 'POST',
  body: formData
});

const result = await response.json();
console.log('File uploaded:', result.id);
```

### Python Example

```python
import requests

url = 'http://localhost:8080/files'
params = {
    'folder': 'documents',
    'uploaded_by': 'python-client'
}

files = {
    'file': open('document.pdf', 'rb')
}

data = {
    'metadata': '{"type": "report", "confidential": true}'
}

response = requests.post(url, params=params, files=files, data=data)
print(response.json())
```

### Go Example

```go
package main

import (
    "bytes"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

func uploadFile(filepath, folder, uploadedBy string) error {
    file, err := os.Open(filepath)
    if err != nil {
        return err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Add file
    part, err := writer.CreateFormFile("file", filepath)
    if err != nil {
        return err
    }
    io.Copy(part, file)

    // Add metadata
    writer.WriteField("metadata", `{"category": "document"}`)
    writer.Close()

    // Create request
    url := "http://localhost:8080/files?folder=" + folder + "&uploaded_by=" + uploadedBy
    req, err := http.NewRequest("POST", url, body)
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

## Method 2: JSON/Base64 Upload (Legacy)

This method accepts base64-encoded file data in a JSON payload. **Not recommended** for most use cases due to:
- Larger payload size (33% overhead from base64 encoding)
- More complex client code
- Less efficient processing

### Endpoint

```
POST /files/upload-json?folder={folder}&uploaded_by={user}
Content-Type: application/json
```

### Request Body

```json
{
  "file": "base64_encoded_file_data",
  "filename": "document.pdf",
  "mime_type": "application/pdf",
  "metadata": {
    "category": "invoice"
  }
}
```

### Example

```bash
# Encode file to base64
FILE_DATA=$(base64 -w 0 document.pdf)

# Upload
curl -X POST "http://localhost:8080/files/upload-json?folder=documents" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"$FILE_DATA\",
    \"filename\": \"document.pdf\",
    \"mime_type\": \"application/pdf\",
    \"metadata\": {\"type\": \"contract\"}
  }"
```

## File Size Limits

- **Minimum**: 1 KB
- **Maximum**: 10 MB

Files outside this range will be rejected with a 400 Bad Request error.

## Supported File Types

The service accepts **all file types**. The MIME type is determined by:

1. **Multipart Upload**: Automatically detected from file content/extension
2. **JSON Upload**: Must be provided in the request

Common MIME types:
- Documents: `application/pdf`, `application/msword`
- Images: `image/jpeg`, `image/png`, `image/gif`
- Text: `text/plain`, `text/csv`
- Archives: `application/zip`, `application/x-tar`
- Default: `application/octet-stream` (binary data)

## Metadata

Metadata is stored as JSON and can contain any custom fields:

```json
{
  "metadata": {
    "category": "invoice",
    "year": 2024,
    "department": "finance",
    "confidential": true,
    "tags": ["important", "urgent"],
    "custom_field": "any value"
  }
}
```

## Unique Filename Generation

The service automatically generates unique filenames:

**Original**: `document.pdf`
**Generated**: `document_65833dffc8a74268bffa6dfb92655d83.pdf`

This ensures:
- No filename collisions
- Original extension preserved
- Files remain identifiable
- Safe characters only

## Error Responses

### File Too Small

```json
{
  "error": "file size too small: minimum is 1KB"
}
```

### File Too Large

```json
{
  "error": "file size too large: maximum is 10MB"
}
```

### No File Provided

```json
{
  "error": "No file provided"
}
```

### Invalid Metadata

```json
{
  "error": "Invalid metadata JSON"
}
```

## Best Practices

### 1. Use Multipart Upload

```bash
# Good - Uses multipart/form-data
curl -X POST "http://localhost:8080/files" \
  -F "file=@document.pdf"

# Avoid - Uses base64 encoding (33% overhead)
curl -X POST "http://localhost:8080/files/upload-json" \
  -H "Content-Type: application/json" \
  -d '{"file": "base64data...", "filename": "doc.pdf"}'
```

### 2. Organize with Folders

```bash
# Organize by type
/files?folder=invoices
/files?folder=contracts
/files?folder=reports

# Organize by user
/files?folder=user-123-documents
/files?folder=user-456-photos

# Organize by date
/files?folder=uploads-2024-11
/files?folder=backups-2024-11-02
```

### 3. Add Meaningful Metadata

```json
{
  "metadata": {
    "category": "invoice",
    "invoice_number": "INV-2024-001",
    "customer_id": "CUST-123",
    "amount": 1500.00,
    "currency": "USD",
    "status": "paid",
    "upload_source": "web_portal"
  }
}
```

### 4. Track Uploader

```bash
# Always include uploaded_by for audit trail
curl -X POST "http://localhost:8080/files?uploaded_by=user@example.com" \
  -F "file=@document.pdf"
```

## Testing

### Create Test File

```bash
# Create a 2KB test file
dd if=/dev/urandom of=test_file.bin bs=2048 count=1

# Upload it
curl -X POST "http://localhost:8080/files?folder=test" \
  -F "file=@test_file.bin"
```

### Batch Upload Script

```bash
#!/bin/bash

for file in /path/to/files/*; do
  echo "Uploading $file..."
  curl -X POST "http://localhost:8080/files?folder=batch-upload" \
    -F "file=@$file" \
    -F "metadata={\"batch_id\": \"$(date +%s)\"}"
done
```

## Integration Examples

### HTML Form

```html
<form action="http://localhost:8080/files?folder=uploads" method="POST" enctype="multipart/form-data">
  <input type="file" name="file" required>
  <input type="hidden" name="metadata" value='{"source": "web_form"}'>
  <button type="submit">Upload</button>
</form>
```

### React/Next.js Component

```tsx
import { useState } from 'react';

export function FileUploader() {
  const [file, setFile] = useState<File | null>(null);

  const handleUpload = async () => {
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);
    formData.append('metadata', JSON.stringify({
      uploadedAt: new Date().toISOString(),
      originalSize: file.size
    }));

    const response = await fetch('http://localhost:8080/files?folder=web-uploads', {
      method: 'POST',
      body: formData
    });

    const result = await response.json();
    console.log('Uploaded:', result.id);
  };

  return (
    <div>
      <input type="file" onChange={(e) => setFile(e.target.files?.[0] || null)} />
      <button onClick={handleUpload}>Upload</button>
    </div>
  );
}
```

## Troubleshooting

### Issue: "No file provided" error

**Cause**: The form field name must be "file"

```bash
# Wrong
curl -F "document=@file.pdf" ...

# Correct
curl -F "file=@file.pdf" ...
```

### Issue: Metadata not saved

**Cause**: Metadata must be valid JSON string

```bash
# Wrong
curl -F 'metadata=not valid json' ...

# Correct
curl -F 'metadata={"key": "value"}' ...
```

### Issue: MIME type is "application/octet-stream"

**Cause**: MIME type auto-detection failed

**Solution**: The file extension might not be recognized. This doesn't affect storage, but you can track the correct type in metadata:

```bash
curl -F "file=@unknown.xyz" \
  -F 'metadata={"actual_mime_type": "application/custom"}'
```

## Summary

**Recommended Approach:**

```bash
curl -X POST "http://localhost:8080/files?folder=my-folder&uploaded_by=user" \
  -F "file=@/path/to/file" \
  -F 'metadata={"key": "value"}'
```

This method is:
- ✅ Standard HTTP multipart/form-data
- ✅ Efficient (no base64 encoding overhead)
- ✅ Simple to implement in any language
- ✅ Supports all file types and sizes (within limits)
- ✅ Automatically detects MIME types
