package pullrequests

import (
	"context"
	"errors"
	"fmt"
	prsDomain "reviewer-assigner/internal/domain/pullrequests"
	"reviewer-assigner/internal/service"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPullRequestRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPostgresPullRequestRepository(
	pool *pgxpool.Pool,
	getter *trmpgx.CtxGetter,
) *PostgresPullRequestRepository {
	return &PostgresPullRequestRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *PostgresPullRequestRepository) GetPullRequestsForReview(
	ctx context.Context,
	userID string,
) ([]prsDomain.PullRequestShort, error) {
	const query = `
	SELECT prs.id, prs.pull_request_id, prs.name, prs.author_id, prs.status FROM pull_requests prs
	JOIN pull_request_reviewers prr on prs.id = prr.pull_request_id
	JOIN users u on u.id = prr.reviewer_id
	WHERE u.user_id = $1
	`

	rows, _ := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, userID)
	pullRequestsDB, err := pgx.CollectRows(rows, pgx.RowToStructByName[PullRequestShortDB])
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	pullRequests := make([]prsDomain.PullRequestShort, 0, len(pullRequestsDB))
	for _, pr := range pullRequestsDB {
		pullRequests = append(pullRequests, *DBShortToDomainPullRequestShort(&pr))
	}

	return pullRequests, nil
}

func (r *PostgresPullRequestRepository) GetByID(
	ctx context.Context,
	pullRequestID string,
) (*prsDomain.PullRequest, error) {
	tx, err := r.getter.DefaultTrOrDB(ctx, r.pool).Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

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
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request reviewers: %w", err)
	}

	pullRequest := DBToDomainPullRequest(pullRequestDB)
	pullRequest.AssignedReviewers = reviewers

	return pullRequest, nil
}

func (r *PostgresPullRequestRepository) Create(
	ctx context.Context,
	pullRequest *prsDomain.PullRequest,
) (string, error) {
	tx, err := r.getter.DefaultTrOrDB(ctx, r.pool).Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const queryInsertPR = `
	INSERT INTO pull_requests (pull_request_id, name, author_id, status) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT DO NOTHING
	RETURNING id, pull_request_id
	`

	var pullRequestSurrogateID int64
	var pullRequestID string
	err = tx.
		QueryRow(
			ctx,
			queryInsertPR,
			pullRequest.ID,
			pullRequest.Name,
			pullRequest.AuthorID,
			pullRequest.Status,
		).
		Scan(&pullRequestSurrogateID, &pullRequestID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", service.ErrPullRequestAlreadyExists
	}
	if err != nil {
		return "", fmt.Errorf("failed to insert pull request: %w", err)
	}

	err = r.insertReviewers(ctx, tx, pullRequest.AssignedReviewers, pullRequestSurrogateID)
	if err != nil {
		return "", fmt.Errorf("failed to insert reviewers: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return pullRequestID, nil
}

func (r *PostgresPullRequestRepository) SetStatusMerged(
	ctx context.Context,
	pullRequestID string,
	mergedAt time.Time,
) error {
	const query = `
	UPDATE pull_requests
	SET status = 'MERGED'::pull_request_status, merged_at = $2
	WHERE pull_request_id = $1
	RETURNING pull_request_id
    `

	var prID string
	err := r.getter.DefaultTrOrDB(ctx, r.pool).
		QueryRow(ctx, query, pullRequestID, mergedAt).
		Scan(&prID)
	if errors.Is(err, pgx.ErrNoRows) {
		return service.ErrPullRequestNotFound
	}
	if err != nil {
		return fmt.Errorf("failed query update: %w", err)
	}

	return nil
}

func (r *PostgresPullRequestRepository) UpdateReviewers(
	ctx context.Context,
	pullRequestID string,
	newReviewerIDs []string,
) error {
	tx, err := r.getter.DefaultTrOrDB(ctx, r.pool).Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const queryDeleteOldReviewers = `
	DELETE FROM pull_request_reviewers
	WHERE pull_request_id = (
		SELECT id FROM pull_requests pr WHERE pr.pull_request_id = $1)
	RETURNING pull_request_id
	`

	var pullRequestSurrogateID int64
	err = tx.QueryRow(ctx, queryDeleteOldReviewers, pullRequestID).Scan(&pullRequestSurrogateID)
	if errors.Is(err, pgx.ErrNoRows) {
		return service.ErrPullRequestNotFound
	}
	if err != nil {
		return fmt.Errorf("failed remove old reviewers: %w", err)
	}

	err = r.insertReviewers(ctx, tx, newReviewerIDs, pullRequestSurrogateID)
	if err != nil {
		return fmt.Errorf("failed to insert new reviewers: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresPullRequestRepository) insertReviewers(
	ctx context.Context,
	txBase pgx.Tx,
	reviewerIDs []string,
	pullRequestSurrogateID int64,
) error {
	tx, err := txBase.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const queryGetSurrogateIDs = `
	SELECT u.id FROM users u
	WHERE u.user_id = ANY($1)
	`

	rows, _ := tx.Query(ctx, queryGetSurrogateIDs, reviewerIDs)
	surrogateReviewerIDs, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return fmt.Errorf("failed to collect reviewer IDs: %w", err)
	}

	if len(surrogateReviewerIDs) == 0 {
		return nil
	}

	const queryInsertReviewers = `
	INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
	VALUES ($1, $2)
	`

	batch := &pgx.Batch{}

	for _, surrogateReviewerID := range surrogateReviewerIDs {
		batch.Queue(queryInsertReviewers, pullRequestSurrogateID, surrogateReviewerID)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to insert batch: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
