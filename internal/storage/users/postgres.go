package users

import (
	"context"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	usersDomain "reviewer-assigner/internal/domain/users"
	"reviewer-assigner/internal/service"
)

type PostgresUserRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPostgresUserRepository(
	pool *pgxpool.Pool,
	getter *trmpgx.CtxGetter,
) *PostgresUserRepository {
	return &PostgresUserRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *PostgresUserRepository) GetUserByID(
	ctx context.Context,
	userID string,
) (*usersDomain.User, error) {
	const query = `
	SELECT u.id, u.user_id, u.name, u.is_active, t.name team_name FROM users u
	JOIN teams t ON t.id = u.team_id
	WHERE user_id = $1
	`

	rows, _ := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, userID)
	userDB, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[UserDB])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to collect user: %w", err)
	}

	return toDomainUser(userDB), nil
}

func (r *PostgresUserRepository) UpdateIsActive(ctx context.Context, user *usersDomain.User) error {
	const query = `
	UPDATE users SET is_active = $1
	WHERE user_id = $2
	RETURNING id
	`

	var surrogateUserID int64
	err := r.getter.DefaultTrOrDB(ctx, r.pool).
		QueryRow(ctx, query, user.IsActive, user.ID).
		Scan(&surrogateUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return service.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
