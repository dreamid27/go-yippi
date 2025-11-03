#!/bin/bash

# Test script for Storage File API

BASE_URL="http://localhost:8080"

echo "=== Testing Storage File API ==="
echo ""

# 1. Upload a test file
echo "1. Uploading a test file..."
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/files?folder=test-folder&uploaded_by=test-user" \
  -H "Content-Type: application/json" \
  -d '{
    "file": "VGhpcyBpcyBhIHRlc3QgZmlsZSBjb250ZW50IGZvciB0ZXN0aW5nIHN0b3JhZ2Ug",
    "filename": "test-document.txt",
    "mime_type": "text/plain",
    "metadata": {
      "description": "Test file for storage service",
      "version": "1.0"
    }
  }')

echo "$UPLOAD_RESPONSE" | jq '.'
FILE_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.id')
echo "Uploaded file ID: $FILE_ID"
echo ""

# 2. Get file by ID
echo "2. Getting file by ID..."
curl -s -X GET "$BASE_URL/files/$FILE_ID" | jq '.'
echo ""

# 3. List all files
echo "3. Listing all files..."
curl -s -X GET "$BASE_URL/files?limit=10" | jq '.'
echo ""

# 4. List files in test-folder
echo "4. Listing files in test-folder..."
curl -s -X GET "$BASE_URL/files?folder=test-folder" | jq '.'
echo ""

# 5. Update metadata
echo "5. Updating file metadata..."
curl -s -X PATCH "$BASE_URL/files/$FILE_ID/metadata" \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "description": "Updated test file",
      "version": "2.0",
      "tags": ["test", "demo"]
    }
  }' | jq '.'
echo ""

# 6. Get file again to see updated metadata
echo "6. Getting file again to verify metadata update..."
curl -s -X GET "$BASE_URL/files/$FILE_ID" | jq '.metadata'
echo ""

# 7. Delete the file
echo "7. Deleting the file..."
curl -s -X DELETE "$BASE_URL/files/$FILE_ID" | jq '.'
echo ""

# 8. Try to get deleted file (should return 404)
echo "8. Trying to get deleted file (should return 404)..."
curl -s -X GET "$BASE_URL/files/$FILE_ID" | jq '.'
echo ""

echo "=== Test completed ==="
