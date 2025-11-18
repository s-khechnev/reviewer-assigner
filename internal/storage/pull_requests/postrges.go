package pull_requests

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
)

type PostgresPullRequestRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPullRequestRepository(pool *pgxpool.Pool) *PostgresPullRequestRepository {
	return &PostgresPullRequestRepository{pool: pool}
}

func (r *PostgresPullRequestRepository) GetPullRequestsForReview(ctx context.Context, userID string) ([]prDomain.PullRequestShort, error) {
	const query = `
	SELECT prs.external_id, prs.name, prs.author_id, prs.status FROM pull_requests prs
	JOIN pull_request_reviewers prr on prs.id = prr.pull_request_id
	WHERE prr.reviewer_id = $1 and prs.status = 'OPEN'
	`

	rows, _ := r.pool.Query(ctx, query, userID)
	pullRequests, err := pgx.CollectRows(rows, pgx.RowToStructByName[prDomain.PullRequestShort])
	if err != nil {
		return nil, err
	}

	return pullRequests, nil
}

func (r *PostgresPullRequestRepository) Get(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error) {
	const query = `
	SELECT prs.external_id, prs.name, prs.author_id, prs.status FROM pull_requests prs
	WHERE prs.external_id = $1
	`

	rows, _ := r.pool.Query(ctx, query, pullRequestID)
	pullRequest, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[prDomain.PullRequest])
	if err != nil {
		return nil, err
	}

	return pullRequest, nil
}

func (r *PostgresPullRequestRepository) Create(ctx context.Context, pullRequest *prDomain.PullRequest) (*prDomain.PullRequest, error) {
	const query = `
	INSERT INTO pull_requests (external_id, name, author_id, status) VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, pullRequest.Id, pullRequest.Name, pullRequest.AuthorId, pullRequest.Status)
	if err != nil {
		return nil, err
	}

	return pullRequest, nil
}

func (r *PostgresPullRequestRepository) Merge(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error) {
	const query = `
    UPDATE pull_requests
    SET status = 'MERGED', merged_at = COALESCE(merged_at, NOW())
    WHERE external_id = $1`

	_, err := r.pool.Exec(ctx, query, pullRequestID)
	if err != nil {
		return nil, err
	}

	return r.Get(ctx, pullRequestID)
}

func (r *PostgresPullRequestRepository) UpdateReviewers(ctx context.Context, pullRequestID string, newReviewerIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var prID int64
	err = tx.QueryRow(ctx, "SELECT id FROM pull_requests WHERE external_id = $1", pullRequestID).Scan(&prID)
	if err != nil {
		return fmt.Errorf("pull request not found: %w", err)
	}

	var newUserIDs []int64
	for _, extID := range newReviewerIDs {
		var userID int64
		err = tx.QueryRow(ctx, "SELECT id FROM users WHERE external_id = $1", extID).Scan(&userID)
		if err != nil {
			return fmt.Errorf("user not found: %s, %w", extID, err)
		}
		newUserIDs = append(newUserIDs, userID)
	}

	const query = `
    WITH current_reviewers AS (
        SELECT reviewer_id FROM pull_request_reviewers WHERE pull_request_id = $1
    ),
    to_delete AS (
        DELETE FROM pull_request_reviewers 
        WHERE pull_request_id = $1 
        AND reviewer_id NOT IN (SELECT unnest($2::bigint[]))
    ),
    to_insert AS (
        INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
        SELECT $1, unnest($2::bigint[])
        ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING
    )
    SELECT 1
    `

	_, err = tx.Exec(ctx, query, prID, newUserIDs)
	if err != nil {
		return fmt.Errorf("failed to update reviewers: %w", err)
	}

	return tx.Commit(ctx)
}
