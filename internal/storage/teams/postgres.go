package teams

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type PostgresTeamRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTeamRepository(pool *pgxpool.Pool) *PostgresTeamRepository {
	return &PostgresTeamRepository{
		pool: pool,
	}
}

// TODO: errors

func (r *PostgresTeamRepository) GetTeam(ctx context.Context, name string) (*teamsDomain.Team, error) {
	rows, err := r.pool.Query(ctx, "SELECT u.external_id id, u.name, u.is_active FROM users u WHERE team_name = $1", name)
	if err != nil {
		return nil, err
	}

	members, err := pgx.CollectRows(rows, pgx.RowToStructByName[teamsDomain.Member])
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, errors.New("team not found")
	}

	return &teamsDomain.Team{
		Name:    name,
		Members: members,
	}, nil
}

func (r *PostgresTeamRepository) CreateTeam(ctx context.Context, teamName string, members []teamsDomain.Member) (*teamsDomain.Team, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "INSERT INTO teams (team_name) VALUES ($1)", teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	query := "INSERT INTO users (external_id, name, team_name) VALUES ($1, $2, $3)"
	batch := &pgx.Batch{}
	for i := range members {
		batch.Queue(query, members[i].ID, members[i].Name, teamName)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return nil, fmt.Errorf("failed to close batch: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &teamsDomain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

func (r *PostgresTeamRepository) UpdateTeam(ctx context.Context, teamName string, members []teamsDomain.Member) (*teamsDomain.Team, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, member := range members {
		batch.Queue(
			"UPDATE users SET name = $1, is_active = $2 WHERE external_id = $3 AND team_name = $4",
			member.Name, member.IsActive, member.ID, teamName,
		)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return nil, fmt.Errorf("failed to close batch: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &teamsDomain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}
