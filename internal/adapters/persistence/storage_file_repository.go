package persistence

import (
	"context"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/storagefile"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/google/uuid"
)

// StorageFileRepositoryImpl implements the StorageFileRepository interface using Ent
type StorageFileRepositoryImpl struct {
	client *ent.Client
}

func NewStorageFileRepository(client *ent.Client) *StorageFileRepositoryImpl {
	return &StorageFileRepositoryImpl{client: client}
}

func (r *StorageFileRepositoryImpl) Create(ctx context.Context, file *entities.StorageFile) error {
	createBuilder := r.client.StorageFile.
		Create().
		SetFilename(file.Filename).
		SetFolder(file.Folder).
		SetOriginalFilename(file.OriginalFilename).
		SetMimeType(file.MimeType).
		SetFileSize(file.FileSize).
		SetFileData(file.FileData).
		SetMetadata(file.Metadata)

	if file.UploadedBy != "" {
		createBuilder = createBuilder.SetUploadedBy(file.UploadedBy)
	}

	created, err := createBuilder.Save(ctx)
	if err != nil {
		return err
	}

	file.ID = created.ID
	file.CreatedAt = created.CreatedAt
	file.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *StorageFileRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.StorageFile, error) {
	found, err := r.client.StorageFile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("StorageFile", id)
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

func (r *StorageFileRepositoryImpl) GetByFilename(ctx context.Context, folder, filename string) (*entities.StorageFile, error) {
	found, err := r.client.StorageFile.
		Query().
		Where(
			storagefile.FolderEQ(folder),
			storagefile.FilenameEQ(filename),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("StorageFile", folder+"/"+filename)
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

func (r *StorageFileRepositoryImpl) ListByFolder(ctx context.Context, folder string) ([]*entities.StorageFile, error) {
	list, err := r.client.StorageFile.
		Query().
		Where(storagefile.FolderEQ(folder)).
		Order(ent.Desc(storagefile.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return r.toEntities(list), nil
}

func (r *StorageFileRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.StorageFile, error) {
	query := r.client.StorageFile.
		Query().
		Order(ent.Desc(storagefile.FieldCreatedAt))

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	return r.toEntities(list), nil
}

func (r *StorageFileRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.StorageFile.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("StorageFile", id)
		}
		return err
	}
	return nil
}

func (r *StorageFileRepositoryImpl) UpdateMetadata(ctx context.Context, id uuid.UUID, metadata map[string]interface{}) error {
	_, err := r.client.StorageFile.
		UpdateOneID(id).
		SetMetadata(metadata).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("StorageFile", id)
		}
		return err
	}
	return nil
}

// Helper methods to convert between Ent and domain entities
func (r *StorageFileRepositoryImpl) toEntity(entFile *ent.StorageFile) *entities.StorageFile {
	return &entities.StorageFile{
		ID:               entFile.ID,
		Filename:         entFile.Filename,
		Folder:           entFile.Folder,
		OriginalFilename: entFile.OriginalFilename,
		MimeType:         entFile.MimeType,
		FileSize:         entFile.FileSize,
		FileData:         entFile.FileData,
		Metadata:         entFile.Metadata,
		UploadedBy:       entFile.UploadedBy,
		CreatedAt:        entFile.CreatedAt,
		UpdatedAt:        entFile.UpdatedAt,
	}
}

func (r *StorageFileRepositoryImpl) toEntities(entFiles []*ent.StorageFile) []*entities.StorageFile {
	files := make([]*entities.StorageFile, 0, len(entFiles))
	for _, f := range entFiles {
		files = append(files, r.toEntity(f))
	}
	return files
}
