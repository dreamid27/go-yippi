package persistence

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/address"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
)

// AddressRepositoryImpl implements the AddressRepository interface using Ent
type AddressRepositoryImpl struct {
	client *ent.Client
}

func NewAddressRepository(client *ent.Client) *AddressRepositoryImpl {
	return &AddressRepositoryImpl{client: client}
}

func (r *AddressRepositoryImpl) Create(ctx context.Context, addr *entities.Address) error {
	builder := r.client.Address.
		Create().
		SetID(addr.ID).
		SetUserID(addr.UserID).
		SetLabel(addr.Label).
		SetRecipientName(addr.RecipientName).
		SetPhone(addr.Phone).
		SetAddressLine1(addr.AddressLine1).
		SetAddressLine2(addr.AddressLine2).
		SetProvinceID(addr.ProvinceID).
		SetProvinceName(addr.ProvinceName).
		SetCityID(addr.CityID).
		SetCityName(addr.CityName).
		SetDistrict(addr.District).
		SetPostalCode(addr.PostalCode).
		SetIsDefault(addr.IsDefault)

	created, err := builder.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domainErrors.NewDuplicateError("Address", "user_id", addr.UserID.String())
		}
		return err
	}

	addr.ID = created.ID
	addr.CreatedAt = created.CreatedAt
	addr.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *AddressRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
	found, err := r.client.Address.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Address", id)
		}
		return nil, err
	}

	// Check if address is deleted
	if found.IsDeleted {
		return nil, domainErrors.NewNotFoundError("Address", id)
	}

	return r.toEntity(found), nil
}

func (r *AddressRepositoryImpl) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Address, error) {
	list, err := r.client.Address.
		Query().
		Where(address.UserID(userID)).
		Where(address.IsDeleted(false)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	addresses := make([]*entities.Address, 0, len(list))
	for _, addr := range list {
		addresses = append(addresses, r.toEntity(addr))
	}

	return addresses, nil
}

func (r *AddressRepositoryImpl) Update(ctx context.Context, addr *entities.Address) error {
	builder := r.client.Address.
		UpdateOneID(addr.ID).
		SetLabel(addr.Label).
		SetRecipientName(addr.RecipientName).
		SetPhone(addr.Phone).
		SetAddressLine1(addr.AddressLine1).
		SetAddressLine2(addr.AddressLine2).
		SetProvinceID(addr.ProvinceID).
		SetProvinceName(addr.ProvinceName).
		SetCityID(addr.CityID).
		SetCityName(addr.CityName).
		SetDistrict(addr.District).
		SetPostalCode(addr.PostalCode).
		SetIsDefault(addr.IsDefault)

	_, err := builder.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Address", addr.ID)
		}
		return err
	}

	return nil
}

func (r *AddressRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Address.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Address", id)
		}
		return err
	}
	return nil
}

func (r *AddressRepositoryImpl) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Address.
		UpdateOneID(id).
		SetIsDeleted(true).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Address", id)
		}
		return err
	}
	return nil
}

func (r *AddressRepositoryImpl) SetDefaultAddress(ctx context.Context, id uuid.UUID) error {
	// Start a transaction for atomic operation
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}

	// Find the address to make default
	defaultAddr, err := tx.Address.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			tx.Rollback()
			return domainErrors.NewNotFoundError("Address", id)
		}
		tx.Rollback()
		return err
	}

	// Update all user's addresses to set is_default=false
	err = tx.Address.
		Update().
		Where(address.UserID(defaultAddr.UserID)).
		SetIsDefault(false).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Set the specified address as default
	_, err = tx.Address.
		UpdateOneID(id).
		SetIsDefault(true).
		Save(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

func (r *AddressRepositoryImpl) GetDefaultAddress(ctx context.Context, userID uuid.UUID) (*entities.Address, error) {
	found, err := r.client.Address.
		Query().
		Where(address.UserID(userID)).
		Where(address.IsDeleted(false)).
		Where(address.IsDefault(true)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Address", "default address for user "+userID.String())
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

// toEntity converts Ent Address to domain entity
func (r *AddressRepositoryImpl) toEntity(addr *ent.Address) *entities.Address {
	return &entities.Address{
		ID:             addr.ID,
		UserID:         addr.UserID,
		Label:          addr.Label,
		RecipientName:  addr.RecipientName,
		Phone:          addr.Phone,
		AddressLine1:   addr.AddressLine1,
		AddressLine2:   addr.AddressLine2,
		ProvinceID:     addr.ProvinceID,
		ProvinceName:   addr.ProvinceName,
		CityID:         addr.CityID,
		CityName:       addr.CityName,
		District:       addr.District,
		PostalCode:     addr.PostalCode,
		IsDefault:      addr.IsDefault,
		IsDeleted:      addr.IsDeleted,
		CreatedAt:      addr.CreatedAt,
		UpdatedAt:      addr.UpdatedAt,
	}
}

// toEnt converts domain Address to Ent entity
func (r *AddressRepositoryImpl) toEnt(addr *entities.Address) *ent.Address {
	// Note: This method is not used in the current implementation but kept for completeness
	// Create would use the builder pattern instead of converting to an existing entity
	return nil
}