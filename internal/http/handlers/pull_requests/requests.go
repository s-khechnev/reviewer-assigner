package pull_requests

type CreatePullRequestRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type MergePullRequestRequest struct {
	ID string `json:"pull_request_id"`
}

type ReassignPullRequestRequest struct {
	ID            string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}
