package pull_requests

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	"reviewer-assigner/internal/service"
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
	SELECT prs.id, prs.pull_request_id, prs.name, prs.author_id, prs.status FROM pull_requests prs
	JOIN pull_request_reviewers prr on prs.id = prr.pull_request_id
	WHERE prr.reviewer_id = $1
`

	rows, _ := r.pool.Query(ctx, queryGetPRs, surrogateUserID)
	pullRequestsDB, err := pgx.CollectRows(rows, pgx.RowToStructByName[PullRequestShortDB])
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	pullRequests := make([]prDomain.PullRequestShort, 0, len(pullRequestsDB))
	for _, pr := range pullRequestsDB {
		pullRequests = append(pullRequests, *DbShortToDomainPullRequestShort(&pr))
	}

	return pullRequests, nil
}

func (r *PostgresPullRequestRepository) Get(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const queryGetPullRequest = `
	SELECT 
	    prs.id, prs.pull_request_id, prs.name, prs.author_id, prs.status, prs.created_at, prs.merged_at 
	FROM pull_requests prs
	WHERE prs.pull_request_id = $1
	`

	rows, _ := tx.Query(ctx, queryGetPullRequest, pullRequestID)
	pullRequestDB, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[PullRequestDB])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, service.ErrPullRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	const queryGetReviewers = `
	SELECT u.user_id FROM users u
	JOIN pull_request_reviewers prr ON u.id = prr.reviewer_id
	WHERE prr.pull_request_id = $1
`

	rows, _ = tx.Query(ctx, queryGetReviewers, pullRequestDB.ID)
	reviewers, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if errors.Is(err, pgx.ErrNoRows) {
		reviewers = []string{}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get pull request reviewers: %w", err)
	}

	pullRequest := DbToDomainPullRequest(pullRequestDB)
	pullRequest.AssignedReviewers = reviewers

	return pullRequest, nil
}

func (r *PostgresPullRequestRepository) Create(ctx context.Context, pullRequest *prDomain.PullRequest) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const queryInsertPR = `
	INSERT INTO pull_requests (pull_request_id, name, author_id, status) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT DO NOTHING
	RETURNING id, pull_request_id
	`

	var pullRequestSurrogateID int64
	var pullRequestID string
	err = tx.
		QueryRow(ctx, queryInsertPR, pullRequest.ID, pullRequest.Name, pullRequest.AuthorID, pullRequest.Status).
		Scan(&pullRequestSurrogateID, &pullRequestID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", service.ErrPullRequestAlreadyExists
	}
	if err != nil {
		return "", fmt.Errorf("failed to insert pull request: %w", err)
	}

	const queryInsertPrReviewers = `
	INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
	VALUES ($1, 
	        (SELECT u.id FROM users u WHERE u.user_id = $2))
	`

	batch := &pgx.Batch{}

	for _, reviewerID := range pullRequest.AssignedReviewers {
		batch.Queue(queryInsertPrReviewers, pullRequestSurrogateID, reviewerID)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return "", fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return pullRequestID, nil
}

func (r *PostgresPullRequestRepository) Merge(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const query = `
    UPDATE pull_requests
    SET status = 'MERGED', merged_at = COALESCE(merged_at, NOW())
    WHERE pull_request_id = $1
    RETURNING *
    `

	rows, _ := tx.Query(ctx, query, pullRequestID)
	pullRequestDB, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[PullRequestDB])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, service.ErrPullRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to merge pull request: %w", err)
	}

	const queryGetReviewers = `
	SELECT u.user_id FROM users u
	JOIN pull_request_reviewers prr ON u.id = prr.reviewer_id
	WHERE prr.pull_request_id = $1
`

	rows, _ = tx.Query(ctx, queryGetReviewers, pullRequestDB.ID)
	reviewers, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if errors.Is(err, pgx.ErrNoRows) {
		reviewers = []string{}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get pull request reviewers: %w", err)
	}

	pullRequest := DbToDomainPullRequest(pullRequestDB)
	pullRequest.AssignedReviewers = reviewers

	return pullRequest, nil
}

func (r *PostgresPullRequestRepository) UpdateReviewers(ctx context.Context, pullRequestID string, newReviewerIDs []string) error {
	const queryDeleteOldReviewers = `
	DELETE FROM pull_request_reviewers
	WHERE pull_request_id = (
		SELECT id FROM pull_requests pr WHERE pr.pull_request_id = $1) 
	`

	if len(newReviewerIDs) == 0 {
		_, err := r.pool.Exec(ctx, queryDeleteOldReviewers, pullRequestID)
		if err != nil {
			return fmt.Errorf("failed to update reviewers: %w", err)
		}

		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	batch.Queue(queryDeleteOldReviewers, pullRequestID)

	const queryInsertNewReviewers = `
	INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
	VALUES (
	        (SELECT id FROM pull_requests pr WHERE pr.pull_request_id = $1),
			(SELECT id FROM users WHERE user_ID = $2)
	)
	`

	for _, newReviewerID := range newReviewerIDs {
		batch.Queue(queryInsertNewReviewers, pullRequestID, newReviewerID)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
