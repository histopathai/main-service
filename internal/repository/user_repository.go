package repository

import (
	"context"
)

const UsersCollection = "users"

type UserRepository struct {
	repo Repository
}

func NewUserRepository(repo Repository) *UserRepository {
	return &UserRepository{
		repo: repo,
	}
}

func (ur *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	return ur.repo.Exists(ctx, UsersCollection, userID)
}
