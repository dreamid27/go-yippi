# Docker Setup Guide

This guide explains how to run the Go-Yippi application using Docker and Docker Compose.

## Prerequisites

- Docker (version 20.10 or higher)
- Docker Compose (version 2.0 or higher)

## Quick Start

### 1. Using Docker Compose (Recommended)

Start all services (PostgreSQL, MinIO, and the application):

```bash
docker-compose up -d
```

This will start:
- **PostgreSQL** on port `5432`
- **MinIO** on port `9000` (API) and `9001` (Console)
- **Go Application** on port `8080`

### 2. Check Service Status

```bash
docker-compose ps
```

### 3. View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f postgres
docker-compose logs -f minio
```

### 4. Stop Services

```bash
docker-compose down
```

To also remove volumes (database and MinIO data):

```bash
docker-compose down -v
```

## Accessing Services

### Application API
- **OpenAPI Docs**: http://localhost:8080/docs
- **Base URL**: http://localhost:8080

### MinIO Console
- **URL**: http://localhost:9001
- **Username**: `minioadmin`
- **Password**: `minioadmin123`

### PostgreSQL
- **Host**: `localhost`
- **Port**: `5432`
- **Database**: `go-test`
- **Username**: `admin`
- **Password**: `adminadmin`

## Configuration

### Environment Variables

You can customize the configuration by creating a `.env` file (copy from `.env.example`):

```bash
cp .env.example .env
```

Then modify the values in `.env` and update the `docker-compose.yml` to use the `.env` file:

```yaml
services:
  app:
    env_file:
      - .env
```

### Default Configuration

The default configuration is optimized for development:

| Service | Variable | Default Value |
|---------|----------|---------------|
| App | SERVER_PORT | 8080 |
| App | SERVER_HOST | 0.0.0.0 |
| PostgreSQL | POSTGRES_USER | admin |
| PostgreSQL | POSTGRES_PASSWORD | adminadmin |
| PostgreSQL | POSTGRES_DB | go-test |
| MinIO | MINIO_ROOT_USER | minioadmin |
| MinIO | MINIO_ROOT_PASSWORD | minioadmin123 |
| MinIO | MINIO_BUCKET_NAME | go-yippi |

## Building the Docker Image

### Build Manually

```bash
docker build -t go-yippi:latest .
```

### Multi-stage Build

The Dockerfile uses a multi-stage build:
1. **Builder stage**: Compiles the Go application with all dependencies
2. **Runtime stage**: Creates a minimal Alpine-based image (~20MB)

## Development Workflow

### 1. Rebuild After Code Changes

```bash
docker-compose up -d --build
```

### 2. Run Tests Inside Container

```bash
docker-compose exec app go test ./...
```

### 3. Access Container Shell

```bash
docker-compose exec app sh
```

## Production Considerations

For production deployment, consider:

1. **Use Docker Secrets** for sensitive data (passwords, keys)
2. **Enable SSL** for MinIO (`MINIO_USE_SSL=true`)
3. **Use managed databases** (AWS RDS, Google Cloud SQL, etc.)
4. **Set up proper backup** for PostgreSQL and MinIO volumes
5. **Configure resource limits** in docker-compose.yml:

```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

6. **Use environment-specific configurations**
7. **Enable health checks** for load balancers
8. **Use reverse proxy** (nginx, Traefik) for SSL termination

## Troubleshooting

### Application won't start

Check if PostgreSQL and MinIO are healthy:
```bash
docker-compose ps
```

View application logs:
```bash
docker-compose logs app
```

### Database connection issues

Ensure PostgreSQL is running and healthy:
```bash
docker-compose exec postgres pg_isready -U admin
```

### MinIO connection issues

Check MinIO logs:
```bash
docker-compose logs minio
```

Verify MinIO is accessible:
```bash
curl http://localhost:9000/minio/health/live
```

### Reset Everything

Stop all services and remove volumes:
```bash
docker-compose down -v
```

Rebuild and restart:
```bash
docker-compose up -d --build
```

## Advanced Usage

### Run Only Dependencies (PostgreSQL + MinIO)

If you want to run the Go app locally but use Docker for dependencies:

1. Comment out the `app` service in `docker-compose.yml`
2. Start only dependencies:
```bash
docker-compose up -d postgres minio
```
3. Run the app locally:
```bash
make run
```

### Custom Network Configuration

The services use a bridge network (`go-yippi-network`). To use a custom network:

```yaml
networks:
  go-yippi-network:
    external: true
    name: my-custom-network
```

## Volume Management

### Backup Volumes

PostgreSQL:
```bash
docker-compose exec postgres pg_dump -U admin go-test > backup.sql
```

MinIO data (using mc CLI):
```bash
docker run --rm --network go-yippi_go-yippi-network \
  -v $(pwd)/minio-backup:/backup \
  minio/mc mirror minio:9000/go-yippi /backup
```

### Restore Volumes

PostgreSQL:
```bash
docker-compose exec -T postgres psql -U admin go-test < backup.sql
```

## Related Documentation

- [Main README](README.md)
- [Architecture Documentation](docs/architecture/overview.md)
- [API Documentation](docs/api/)
