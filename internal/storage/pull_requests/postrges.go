package pull_requests

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
)

type PostgresPullRequestRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPullRequestRepository(pool *pgxpool.Pool) *PostgresPullRequestRepository {
	return &PostgresPullRequestRepository{
		pool: pool,
	}
}

func (r *PostgresPullRequestRepository) GetPullRequestsForReview(ctx context.Context, userID string) ([]prDomain.PullRequestShort, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const queryGetSurrogateUserID = `
	SELECT u.id FROM users u
	WHERE u.user_id = $1
`

	var surrogateUserID int64
	err = tx.QueryRow(ctx, queryGetSurrogateUserID, userID).Scan(&surrogateUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Is it ok?
		return []prDomain.PullRequestShort{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	const queryGetPRs = `
	SELECT prs.pull_request_id, prs.name, prs.author_id, prs.status FROM pull_requests prs
	JOIN pull_request_reviewers prr on prs.id = prr.pull_request_id
	WHERE prr.reviewer_id = $1 and prs.status = 'OPEN'
`

	rows, _ := r.pool.Query(ctx, queryGetPRs, surrogateUserID)
	pullRequestsDB, err := pgx.CollectRows(rows, pgx.RowToStructByName[PullRequestDB])
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	pullRequests := make([]prDomain.PullRequestShort, 0, len(pullRequestsDB))
	for _, pr := range pullRequestsDB {
		pullRequests = append(pullRequests, *DbToDomainPullRequestShort(&pr))
	}

	return pullRequests, nil
}

func (r *PostgresPullRequestRepository) Get(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error) {
	const query = `
	SELECT prs.pull_request_id, prs.name, prs.author_id, prs.status FROM pull_requests prs
	WHERE prs.pull_request_id = $1
	`

	rows, _ := r.pool.Query(ctx, query, pullRequestID)
	pullRequestDB, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[PullRequestDB])
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	return DbToDomainPullRequest(pullRequestDB), nil
}

func (r *PostgresPullRequestRepository) Create(ctx context.Context, pullRequest *prDomain.PullRequest) (string, error) {
	const query = `
	INSERT INTO pull_requests (pull_request_id, name, author_id, status) VALUES ($1, $2, $3, $4)
	RETURNING pull_request_id
	`

	var pullRequestID string
	err := r.pool.QueryRow(ctx, query, pullRequest.ID, pullRequest.Name, pullRequest.AuthorID, pullRequest.Status).Scan(&pullRequestID)
	if err != nil {
		return "", fmt.Errorf("failed to insert pull request: %w", err)
	}

	return pullRequestID, nil
}

func (r *PostgresPullRequestRepository) Merge(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error) {
	const query = `
    UPDATE pull_requests
    SET status = 'MERGED', merged_at = COALESCE(merged_at, NOW())
    WHERE pull_request_id = $1
    RETURNING pull_request_id, name, author_id, status, created_at, merged_at
    `

	rows, _ := r.pool.Query(ctx, query, pullRequestID)
	pullRequestDB, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[PullRequestDB])
	if err != nil {
		return nil, fmt.Errorf("failed to merge pull request: %w", err)
	}

	return DbToDomainPullRequest(pullRequestDB), nil
}

func (r *PostgresPullRequestRepository) UpdateReviewers(ctx context.Context, pullRequestID string, newReviewerIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var prSurrogateID int64
	err = tx.QueryRow(ctx, "SELECT id FROM pull_requests WHERE pull_request_id = $1", pullRequestID).Scan(&prSurrogateID)
	if err != nil {
		return fmt.Errorf("pull request not found: %w", err)
	}

	var reviewerSurrogateIDs []int64
	for _, reviewerID := range newReviewerIDs {
		var surrogateID int64
		err = tx.QueryRow(ctx, "SELECT id FROM users WHERE user_id = $1", reviewerID).Scan(&surrogateID)
		if err != nil {
			return fmt.Errorf("user not found: %s, %w", reviewerID, err)
		}
		reviewerSurrogateIDs = append(reviewerSurrogateIDs, surrogateID)
	}

	const query = `
    INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
    VALUES ($1, $2)
    ON CONFLICT DO UPDATE
    SET reviewer_id = $2
    `

	batch := &pgx.Batch{}
	for _, reviewerSurrogateID := range reviewerSurrogateIDs {
		batch.Queue(query, prSurrogateID, reviewerSurrogateID)
	}
	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
