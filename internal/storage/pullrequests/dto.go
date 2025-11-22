package pullrequests

import (
	prsDomain "reviewer-assigner/internal/domain/pullrequests"
	"time"
)

type PullRequestShortDB struct {
	ID            int64              `db:"id"`
	PullRequestID string             `db:"pull_request_id"`
	Name          string             `db:"name"`
	AuthorID      string             `db:"author_id"`
	Status        prsDomain.StatusPR `db:"status"`
}

type PullRequestDB struct {
	PullRequestShortDB

	CreatedAt *time.Time `db:"created_at"`
	MergedAt  *time.Time `db:"merged_at"`
}

func DBShortToDomainPullRequestShort(d *PullRequestShortDB) *prsDomain.PullRequestShort {
	return &prsDomain.PullRequestShort{
		ID:       d.PullRequestID,
		Name:     d.Name,
		AuthorID: d.AuthorID,
		Status:   d.Status,
	}
}

func DBToDomainPullRequest(d *PullRequestDB) *prsDomain.PullRequest {
	return &prsDomain.PullRequest{
		PullRequestShort: prsDomain.PullRequestShort{
			ID:       d.PullRequestID,
			Name:     d.Name,
			AuthorID: d.AuthorID,
			Status:   d.Status,
		},
		CreatedAt: d.CreatedAt,
		MergedAt:  d.MergedAt,
	}
}
