package pull_requests

import "time"

type StatusPR string

const (
	StatusOpen   StatusPR = "OPEN"
	StatusMerged StatusPR = "MERGED"
)

type PullRequestShort struct {
	ID       string
	Name     string
	AuthorID string
	Status   StatusPR
}

type PullRequest struct {
	PullRequestShort
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
