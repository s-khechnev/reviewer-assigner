package teams

import (
	"context"
	"errors"
	"fmt"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/service"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTeamRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPostgresTeamRepository(
	pool *pgxpool.Pool,
	getter *trmpgx.CtxGetter,
) *PostgresTeamRepository {
	return &PostgresTeamRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *PostgresTeamRepository) GetTeamByName(
	ctx context.Context,
	teamName string,
) (*teamsDomain.Team, error) {
	const query = `
	SELECT u.id, u.user_id, u.name, u.is_active FROM teams t
	JOIN users u ON t.id = u.team_id
	WHERE t.name = $1
	`

	rows, _ := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, teamName)
	membersDB, err := pgx.CollectRows(rows, pgx.RowToStructByName[MemberDB])
	if err != nil {
		return nil, fmt.Errorf("failed to collect team members: %w", err)
	}

	if len(membersDB) == 0 {
		return nil, service.ErrTeamNotFound
	}

	members := make([]teamsDomain.Member, 0, len(membersDB))
	for _, member := range membersDB {
		members = append(members, *DBToDomainMember(&member))
	}

	return &teamsDomain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

func (r *PostgresTeamRepository) SaveTeam(
	ctx context.Context,
	teamName string,
	members []teamsDomain.Member,
) (int64, error) {
	tx, err := r.getter.DefaultTrOrDB(ctx, r.pool).Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

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
	INSERT INTO users (user_id, name, team_id, is_active)
	VALUES ($1, $2, $3, $4)
	`

	batch := &pgx.Batch{}
	for _, member := range members {
		batch.Queue(queryInsertMember, member.ID, member.Name, teamID, member.IsActive)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return 0, fmt.Errorf("failed to batch insert members: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return teamID, nil
}

func (r *PostgresTeamRepository) UpdateMembers(
	ctx context.Context,
	teamName string,
	updatedMembers []teamsDomain.Member,
) error {
	tx, err := r.getter.DefaultTrOrDB(ctx, r.pool).Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const queryTeamID = `
	SELECT id FROM teams WHERE name = $1
	`

	var teamID int64
	err = tx.QueryRow(ctx, queryTeamID, teamName).Scan(&teamID)
	if errors.Is(err, pgx.ErrNoRows) {
		return service.ErrTeamNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	const query = `
	UPDATE users
	SET name = $1, is_active = $2 
	WHERE user_id = $3 AND team_id = $4
	`

	batch := &pgx.Batch{}
	for _, member := range updatedMembers {
		batch.Queue(query,
			member.Name, member.IsActive, member.ID, teamID,
		)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to batch update updatedMembers: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
