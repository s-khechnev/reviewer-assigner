package pull_requests

import (
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	"time"
)

type PullRequestDb struct {
	ID            int64             `db:"id"`
	PullRequestID string            `db:"pull_request_id"`
	Name          string            `db:"name"`
	AuthorID      string            `db:"author_id"`
	Status        prDomain.StatusPR `db:"status"`
	CreatedAt     *time.Time        `db:"created_at"`
	MergedAt      *time.Time        `db:"merged_at"`
}

func DbToDomainPullRequestShort(d *PullRequestDb) *prDomain.PullRequestShort {
	return &prDomain.PullRequestShort{
		ID:       d.PullRequestID,
		Name:     d.Name,
		AuthorID: d.AuthorID,
		Status:   d.Status,
	}
}

func DbToDomainPullRequest(d *PullRequestDb) *prDomain.PullRequest {
	return &prDomain.PullRequest{
		PullRequestShort: prDomain.PullRequestShort{
			ID:       d.PullRequestID,
			Name:     d.Name,
			AuthorID: d.AuthorID,
			Status:   d.Status,
		},
		CreatedAt: d.CreatedAt,
		MergedAt:  d.MergedAt,
	}
}
