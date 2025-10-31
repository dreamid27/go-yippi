package persistence

import (
	"context"

	"example.com/go-yippi/ent"
	"example.com/go-yippi/internal/domain/entities"
)

// UserRepositoryImpl implements the UserRepository interface using Ent
type UserRepositoryImpl struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepositoryImpl {
	return &UserRepositoryImpl{client: client}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entities.User) error {
	created, err := r.client.User.
		Create().
		SetName(user.Name).
		SetAge(user.Age).
		Save(ctx)
	if err != nil {
		return err
	}

	user.ID = created.ID
	return nil
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id int) (*entities.User, error) {
	found, err := r.client.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &entities.User{
		ID:    found.ID,
		Name:  found.Name,
		Age: found.Age,
	}, nil
}

func (r *UserRepositoryImpl) List(ctx context.Context) ([]*entities.User, error) {
	list, err := r.client.User.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]*entities.User, 0, len(list))
	for _, u := range list {
		users = append(users, &entities.User{
			ID:    u.ID,
			Name:  u.Name,
			Age: u.Age,
		})
	}

	return users, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *entities.User) error {
	_, err := r.client.User.
		UpdateOneID(user.ID).
		SetName(user.Name).
		SetAge(user.Age).
		Save(ctx)
	return err
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.client.User.DeleteOneID(id).Exec(ctx)
}
