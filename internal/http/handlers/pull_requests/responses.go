package pull_requests

import (
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	"time"
)

type CreatePullRequestResponse struct {
	PullRequestResponse `json:"pr"`
}

type MergePullRequestResponse struct {
	PullRequestResponse `json:"pr"`
}

type ReassignPullRequestResponse struct {
	PullRequestResponse `json:"pr"`
	ReplacedBy          string `json:"replaced_by"`
}

type PullRequestResponse struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	MergedAt          *time.Time `json:"merged_at,omitempty"`
}

func domainToCreatePullRequestResponse(pr *prDomain.PullRequest) *CreatePullRequestResponse {
	return &CreatePullRequestResponse{
		PullRequestResponse: *domainToPullRequestResponse(pr),
	}
}

func domainToMergePullRequestResponse(pr *prDomain.PullRequest) *MergePullRequestResponse {
	return &MergePullRequestResponse{
		PullRequestResponse: *domainToPullRequestResponse(pr),
	}
}

func domainToReassignPullRequestResponse(pr *prDomain.PullRequest, replacedBy string) *ReassignPullRequestResponse {
	return &ReassignPullRequestResponse{
		PullRequestResponse: *domainToPullRequestResponse(pr),
		ReplacedBy:          replacedBy,
	}
}

func domainToPullRequestResponse(pr *prDomain.PullRequest) *PullRequestResponse {
	return &PullRequestResponse{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}
