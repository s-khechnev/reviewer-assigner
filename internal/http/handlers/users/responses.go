package users

import (
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	usersDomain "reviewer-assigner/internal/domain/users"
)

type UserResponse struct {
	UserID   string `json:"user_id"`
	Name     string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type GetReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestResponse `json:"pull_requests"`
}

type PullRequestResponse struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

func domainToGetReviewResponse(userID string, prs []prDomain.PullRequestShort) GetReviewResponse {
	prsResponse := make([]PullRequestResponse, 0, len(prs))
	for _, pr := range prs {
		prsResponse = append(prsResponse, PullRequestResponse{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   string(pr.Status),
		})
	}

	return GetReviewResponse{
		UserID:       userID,
		PullRequests: prsResponse,
	}
}

func domainToUserResponse(u *usersDomain.User) UserResponse {
	return UserResponse{
		UserID:   u.ID,
		Name:     u.Name,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}
