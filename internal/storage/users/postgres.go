package users

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	usersDomain "reviewer-assigner/internal/domain/users"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Get(ctx context.Context, userID string) (*usersDomain.User, error) {

	return nil, nil
}
