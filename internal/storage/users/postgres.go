package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	usersDomain "reviewer-assigner/internal/domain/users"
	"reviewer-assigner/internal/service"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		pool: pool,
	}
}

func (r *PostgresUserRepository) Get(ctx context.Context, userID string) (*usersDomain.User, error) {
	const query = `
	SELECT u.user_id, u.name, u.is_active, t.name FROM users u
	JOIN teams t ON t.id = u.team_id
	WHERE user_id = $1
	`

	var user usersDomain.User
	err := r.pool.
		QueryRow(ctx, query, userID).
		Scan(&user.ID, &user.Name, &user.IsActive, &user.TeamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*usersDomain.User, error) {
	const query = `
    UPDATE users
    SET is_active = $1
    WHERE user_id = $2
    RETURNING user_id, name, is_active, (SELECT name FROM teams WHERE id = team_id) as team_name
    `

	var user usersDomain.User
	err := r.pool.QueryRow(ctx, query, isActive, userID).Scan(&user.ID, &user.Name, &user.IsActive, &user.TeamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}
