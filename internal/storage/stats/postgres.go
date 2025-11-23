package stats

import (
	"context"
	"fmt"
	"reviewer-assigner/internal/http/handlers/stats"
	"strings"

	"github.com/jackc/pgx/v5"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStatsRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPostgresStatsRepository(
	pool *pgxpool.Pool,
	getter *trmpgx.CtxGetter,
) *PostgresStatsRepository {
	return &PostgresStatsRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *PostgresStatsRepository) GetStatsReviewersAssignments(
	ctx context.Context, status string, activeOnly bool,
) ([]stats.UserAssignment, error) {
	// TODO: add pagination
	const queryBase = `
	SELECT
		u.user_id,
		u.name as username,
		COUNT(prr.pull_request_id) as assignment_count
	FROM users u
	LEFT JOIN pull_request_reviewers prr ON u.id = prr.reviewer_id
	LEFT JOIN pull_requests pr ON prr.pull_request_id = pr.id
	`

	var queryBuilder strings.Builder
	queryBuilder.WriteString(queryBase)

	var args []any
	var whereConditions []string
	if status != "" {
		whereConditions = append(whereConditions, "pr.status = $1::pull_request_status")
		args = append(args, status)
	}
	if activeOnly {
		whereConditions = append(whereConditions, "u.is_active = true")
	}

	if len(whereConditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(whereConditions, " AND "))
	}

	queryBuilder.WriteString(`
        GROUP BY u.user_id, u.name
		HAVING count(prr.pull_request_id) > 0
        ORDER BY assignment_count DESC
    `)

	query := queryBuilder.String()

	rows, _ := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, args...)
	userAssignmentDBs, err := pgx.CollectRows(rows, pgx.RowToStructByName[UserAssignmentDB])
	if err != nil {
		return nil, fmt.Errorf("failed to collect assignments: %w", err)
	}

	userAssignments := make([]stats.UserAssignment, 0, len(userAssignmentDBs))
	for _, userAssignmentDB := range userAssignmentDBs {
		userAssignments = append(userAssignments, DBToDomainUserAssignment(&userAssignmentDB))
	}

	return userAssignments, nil
}
