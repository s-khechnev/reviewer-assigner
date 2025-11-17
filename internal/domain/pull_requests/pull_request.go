package pull_requests

import "time"

type StatusPR string

var (
	StatusOpen   StatusPR = "OPEN"
	StatusMerged StatusPR = "MERGED"
)

type PullRequestShort struct {
	Id       string
	Name     string
	AuthorId string
	Status   StatusPR
}

type PullRequest struct {
	PullRequestShort
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
