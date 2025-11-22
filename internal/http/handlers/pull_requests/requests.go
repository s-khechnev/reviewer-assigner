package pull_requests

type CreatePullRequestRequest struct {
	ID       string `json:"pull_request_id" validate:"required"`
	Name     string `json:"pull_request_name" validate:"required"`
	AuthorID string `json:"author_id" validate:"required"`
}

type MergePullRequestRequest struct {
	ID string `json:"pull_request_id" validate:"required"`
}

type ReassignPullRequestRequest struct {
	ID            string `json:"pull_request_id" validate:"required"`
	OldReviewerID string `json:"old_reviewer_id" validate:"required"`
}
