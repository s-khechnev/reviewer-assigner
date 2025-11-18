package users

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
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
	const query = `
	SELECT u.external_id, u.name, u.is_active, u.team_name FROM users u WHERE external_id = $1
	`

	rows, _ := r.pool.Query(ctx, query, userID)
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[usersDomain.User])
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *PostgresUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*usersDomain.User, error) {
	const query = `
        WITH updated_user AS (
            UPDATE users 
            SET is_active = $1 
            WHERE external_id = $2
            RETURNING external_id, name, is_active, team_name
        )
        SELECT external_id, name, is_active, team_name
        FROM updated_user
    `

	rows, _ := r.pool.Query(ctx, query, userID, isActive)
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[usersDomain.User])
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
