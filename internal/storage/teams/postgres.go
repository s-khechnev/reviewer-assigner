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
	const query = `
	SELECT u.user_id, u.name, u.is_active FROM users u
	JOIN teams t ON t.id = u.team_id
	WHERE t.name = $1
`

	rows, err := r.pool.Query(ctx, query, name)
	if err != nil {
		return nil, err
	}

	membersDB, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MemberDb])
	if err != nil {
		return nil, err
	}

	if len(membersDB) == 0 {
		return nil, errors.New("team not found")
	}

	members := make([]teamsDomain.Member, 0, len(membersDB))
	for _, member := range membersDB {
		members = append(members, *DbToDomainMember(member))
	}

	return &teamsDomain.Team{
		Name:    name,
		Members: members,
	}, nil
}

func (r *PostgresTeamRepository) SaveTeam(ctx context.Context, teamName string, members []teamsDomain.Member) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const queryInsertTeam = `
	INSERT INTO teams (name) VALUES ($1)
	RETURNING id
`

	var teamID int64
	err = tx.QueryRow(ctx, queryInsertTeam, teamName).Scan(&teamID)
	if err != nil {
		return 0, fmt.Errorf("failed to create team: %w", err)
	}

	const queryInsertMember = `
	INSERT INTO users (user_id, name, team_id) VALUES ($1, $2, $3)
	`

	batch := &pgx.Batch{}
	for i := range members {
		batch.Queue(queryInsertMember, members[i].ID, members[i].Name, teamName)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return 0, fmt.Errorf("failed to close batch: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return teamID, nil
}

func (r *PostgresTeamRepository) UpdateTeam(ctx context.Context, teamName string, members []teamsDomain.Member) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const queryTeamID = `SELECT id FROM teams WHERE name = $4`
	var teamID int64
	err = tx.QueryRow(ctx, queryTeamID, teamName).Scan(&teamID)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	const query = `
	UPDATE users 
	SET name = $1, is_active = $2 
	WHERE user_id = $3 AND team_id = $4
`

	batch := &pgx.Batch{}
	for _, member := range members {
		batch.Queue(query,
			member.Name, member.IsActive, member.ID, teamID,
		)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to close batch: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
