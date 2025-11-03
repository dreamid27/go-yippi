#!/bin/bash

# Test script for MinIO Storage Backend

BASE_URL="http://localhost:8080"

echo "=== Testing MinIO Storage Backend ==="
echo ""

# Generate test data
TEST_DATA=$(python3 -c "import base64; print(base64.b64encode(b'This is a test file stored in MinIO object storage!' * 50).decode())")

# 1. Upload a test file
echo "1. Uploading a test file to MinIO..."
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/files?folder=test-minio&uploaded_by=minio-user" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"$TEST_DATA\",
    \"filename\": \"minio-test-document.txt\",
    \"mime_type\": \"text/plain\",
    \"metadata\": {
      \"storage\": \"minio\",
      \"test\": true,
      \"description\": \"Testing MinIO integration\"
    }
  }")

echo "$UPLOAD_RESPONSE" | jq '.'
FILE_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.id')
FILENAME=$(echo "$UPLOAD_RESPONSE" | jq -r '.filename')
echo "Uploaded file ID: $FILE_ID"
echo "Generated filename: $FILENAME"
echo ""

# 2. Get file by ID
echo "2. Getting file by ID from MinIO..."
curl -s -X GET "$BASE_URL/files/$FILE_ID" | jq '{id, filename, folder, original_filename, mime_type, file_size, metadata, uploaded_by}'
echo ""

# 3. Get file by path
echo "3. Getting file by path (folder/filename)..."
curl -s -X GET "$BASE_URL/files/test-minio/$FILENAME" | jq '{id, filename, folder, file_size}'
echo ""

# 4. List files in folder
echo "4. Listing files in test-minio folder..."
curl -s -X GET "$BASE_URL/files?folder=test-minio" | jq '.'
echo ""

# 5. Update metadata
echo "5. Updating file metadata..."
curl -s -X PATCH "$BASE_URL/files/$FILE_ID/metadata" \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "storage": "minio",
      "test": true,
      "description": "Updated metadata via MinIO backend",
      "version": "2.0",
      "tags": ["minio", "object-storage", "test"]
    }
  }' | jq '.'
echo ""

# 6. Verify metadata update
echo "6. Verifying metadata update..."
curl -s -X GET "$BASE_URL/files/$FILE_ID" | jq '.metadata'
echo ""

# 7. Upload another file
echo "7. Uploading another file..."
TEST_DATA2=$(python3 -c "import base64; print(base64.b64encode(b'Second test file in MinIO!' * 40).decode())")
UPLOAD_RESPONSE2=$(curl -s -X POST "$BASE_URL/files?folder=test-minio&uploaded_by=minio-user" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"$TEST_DATA2\",
    \"filename\": \"second-file.txt\",
    \"mime_type\": \"text/plain\",
    \"metadata\": {}
  }")
FILE_ID2=$(echo "$UPLOAD_RESPONSE2" | jq -r '.id')
echo "Second file ID: $FILE_ID2"
echo ""

# 8. List all files
echo "8. Listing all files in test-minio folder..."
curl -s -X GET "$BASE_URL/files?folder=test-minio&limit=10" | jq '.files | length'
echo ""

# 9. Delete first file
echo "9. Deleting first file..."
curl -s -X DELETE "$BASE_URL/files/$FILE_ID" | jq '.'
echo ""

# 10. Verify deletion
echo "10. Verifying file was deleted (should return 404)..."
curl -s -X GET "$BASE_URL/files/$FILE_ID" | jq '.'
echo ""

# 11. List files again (should show only second file)
echo "11. Listing files after deletion..."
curl -s -X GET "$BASE_URL/files?folder=test-minio" | jq '.total'
echo ""

# Cleanup - delete second file
echo "Cleanup: Deleting second file..."
curl -s -X DELETE "$BASE_URL/files/$FILE_ID2" | jq '.'

echo ""
echo "=== MinIO Integration Test Completed ==="
