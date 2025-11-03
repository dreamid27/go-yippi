# MinIO Integration Guide

The Storage File Service now supports **MinIO** as an object storage backend alongside PostgreSQL database storage.

## Architecture

The service uses the **hexagonal architecture** pattern with a pluggable storage backend:

```
Domain Layer (StorageFileRepository interface)
        ↓
Application Layer (StorageFileService)
        ↓
Adapter Layer (choose one):
├── StorageFileRepositoryImpl (PostgreSQL - BYTEA storage)
└── StorageFileMinIORepository (MinIO - S3-compatible object storage)
```

## Configuration

### Environment Variables

```bash
# Storage backend selection
STORAGE_BACKEND=database  # Options: "database" or "minio" (default: "database")

# MinIO configuration (only needed when STORAGE_BACKEND=minio)
MINIO_ENDPOINT=localhost:9000         # MinIO server endpoint
MINIO_ACCESS_KEY=minioadmin           # Access key ID
MINIO_SECRET_KEY=minioadmin123        # Secret access key
MINIO_USE_SSL=false                   # Use HTTPS (default: false)
MINIO_BUCKET_NAME=storage-files       # Bucket name (default: "storage-files")
```

### Docker Compose MinIO Setup

```yaml
services:
  minio:
    image: minio/minio:latest
    container_name: minio
    restart: unless-stopped
    ports:
      - "9000:9000"   # API (S3)
      - "9001:9001"   # Console UI
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    volumes:
      - minio-data:/data
    command: server /data --console-address ":9001"

volumes:
  minio-data:
```

## Running with Different Backends

### Run with PostgreSQL (Default)

```bash
# Uses database storage
make run

# Or explicitly
STORAGE_BACKEND=database make run
```

### Run with MinIO

```bash
# Uses MinIO object storage
STORAGE_BACKEND=minio make run
```

## Storage Backend Comparison

| Feature | PostgreSQL | MinIO |
|---------|-----------|-------|
| Storage Type | Database BYTEA column | S3-compatible object storage |
| Best For | Small files, transactional data | Large files, scalable storage |
| Scalability | Limited by database size | Highly scalable |
| Backup | Database backup | Object storage backup |
| Performance | Fast for small files | Optimized for large files |
| Cost | Database storage cost | Cheaper for large files |
| Query Support | Full SQL queries | Metadata-based search |

## MinIO Implementation Details

### Object Naming Convention

Files are stored in MinIO using the following path structure:

```
{bucket}/{folder}/{filename}
```

Example:
```
storage-files/user-avatars/profile_abc123def456.jpg
storage-files/documents/report_xyz789.pdf
```

### Metadata Storage

MinIO stores file metadata as S3 user metadata (headers):

```
x-amz-meta-id: UUID
x-amz-meta-folder: folder-name
x-amz-meta-original-filename: original.txt
x-amz-meta-uploaded-by: user@example.com
x-amz-meta-custom-metadata: {"key":"value",...}
```

### Automatic Bucket Creation

The application automatically creates the bucket if it doesn't exist on startup:

```
2025/11/02 20:08:44 Using MinIO storage backend
2025/11/02 20:08:44 Created bucket: storage-files
```

## Code Examples

### Creating Storage Repository

The main.go automatically selects the correct repository based on `STORAGE_BACKEND`:

```go
var storageFileRepo ports.StorageFileRepository

switch cfg.Storage.Backend {
case "minio":
    // Initialize MinIO client
    minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey, ""),
        Secure: cfg.MinIO.UseSSL,
    })

    // Create bucket if not exists
    exists, _ := minioClient.BucketExists(ctx, cfg.MinIO.BucketName)
    if !exists {
        minioClient.MakeBucket(ctx, cfg.MinIO.BucketName, minio.MakeBucketOptions{})
    }

    storageFileRepo = persistence.NewStorageFileMinIORepository(minioClient, cfg.MinIO.BucketName)

case "database":
    storageFileRepo = persistence.NewStorageFileRepository(client)
}
```

### Using the Service

The API remains **exactly the same** regardless of backend:

```bash
# Upload (works with both backends)
curl -X POST "http://localhost:8080/files?folder=documents" \
  -H "Content-Type: application/json" \
  -d '{
    "file": "base64_encoded_data",
    "filename": "document.pdf",
    "mime_type": "application/pdf"
  }'

# Download (works with both backends)
curl -X GET "http://localhost:8080/files/{id}"

# List (works with both backends)
curl -X GET "http://localhost:8080/files?folder=documents"
```

## Testing

### Test with MinIO

```bash
# Start MinIO and PostgreSQL
docker-compose up -d

# Run application with MinIO backend
STORAGE_BACKEND=minio make run

# Run test script
./test_minio_api.sh
```

### Access MinIO Console

Open http://localhost:9001 in your browser:

- Username: `minioadmin`
- Password: `minioadmin123`

You can visually browse, download, and manage files through the console.

## Advantages of MinIO Backend

1. **Scalability**: MinIO scales horizontally across multiple servers
2. **Cost-Effective**: Cheaper storage for large files
3. **S3 Compatible**: Works with AWS S3 clients and tools
4. **High Performance**: Optimized for object storage workloads
5. **Cloud-Ready**: Easy to migrate to AWS S3, Google Cloud Storage, etc.
6. **Better for Large Files**: More efficient than database BYTEA storage
7. **Separation of Concerns**: File storage separate from transactional database

## When to Use Each Backend

### Use PostgreSQL (database) when:

- ✅ Files are small (< 1MB)
- ✅ Need ACID transactions with file data
- ✅ Simpler deployment (no extra services)
- ✅ Need complex queries on file metadata
- ✅ Development/testing environment

### Use MinIO (object storage) when:

- ✅ Files are large (> 1MB)
- ✅ Need scalable storage
- ✅ High throughput requirements
- ✅ Production environment
- ✅ Planning to scale horizontally
- ✅ Need S3-compatible storage
- ✅ Cost optimization for storage

## Migration Between Backends

To migrate from one backend to another:

1. Export files from current backend
2. Change `STORAGE_BACKEND` environment variable
3. Import files to new backend
4. Update application configuration

Note: A migration script can be created to automate this process.

## Troubleshooting

### MinIO Connection Issues

```bash
# Check if MinIO is running
docker ps | grep minio

# Check MinIO logs
docker logs minio

# Test MinIO connectivity
curl http://localhost:9000/minio/health/live
```

### Bucket Not Created

The application automatically creates the bucket. If issues persist:

```bash
# Manually create bucket using mc (MinIO Client)
mc alias set local http://localhost:9000 minioadmin minioadmin123
mc mb local/storage-files
```

### File Not Found Errors

- Ensure bucket exists
- Check file path format (folder/filename)
- Verify metadata is correctly stored
- Check MinIO console for actual file locations

## Performance Tuning

### For MinIO

1. **Use Multiple Drives**: MinIO performs best with multiple drives
2. **Enable Caching**: Configure read/write caching
3. **Tune Network**: Use high-speed network for distributed setups
4. **Adjust Workers**: Configure concurrent uploads/downloads

### Configuration Example

```go
// Increase concurrent connections
transport := &http.Transport{
    MaxIdleConns:       100,
    MaxIdleConnsPerHost: 100,
}

minioClient, err := minio.New(endpoint, &minio.Options{
    Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
    Secure:    useSSL,
    Transport: transport,
})
```

## Security Best Practices

1. **Use HTTPS**: Set `MINIO_USE_SSL=true` in production
2. **Strong Credentials**: Change default minioadmin credentials
3. **Access Policies**: Configure bucket policies for access control
4. **Encryption**: Enable encryption at rest and in transit
5. **Network Isolation**: Run MinIO in private network
6. **Regular Backups**: Configure automated backups

## Future Enhancements

- [ ] Add AWS S3 backend support
- [ ] Add Google Cloud Storage backend support
- [ ] Add Azure Blob Storage backend support
- [ ] Implement migration tool between backends
- [ ] Add file versioning support
- [ ] Add file compression option
- [ ] Add CDN integration for public files
- [ ] Add file preview generation
