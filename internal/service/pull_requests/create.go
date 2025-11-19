package pull_requests

import (
	"context"
	"fmt"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/service"
	"time"
)

// TODO: transactions, errors

func (s *PullRequestService) Create(ctx context.Context, prID, prName, authorID string) (*prDomain.PullRequest, error) {
	author, err := s.userRepo.Get(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", service.ErrUserNotFound)
	}

	activeMembers, err := s.teamRepo.GetActiveMembers(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get active members: %w", err)
	}

	activeMembersExcludeAuthor := make([]teamsDomain.Member, 0, len(activeMembers))
	for _, member := range activeMembers {
		if member.ID != author.ID {
			activeMembersExcludeAuthor = append(activeMembersExcludeAuthor, member)
		}
	}

	reviewers, err := s.reviewerPicker.Pick(activeMembersExcludeAuthor)
	if err != nil {
		return nil, fmt.Errorf("failed to pick reviewers: %w", err)
	}

	reviewersIDs := make([]string, 0, len(reviewers))
	for i := range reviewers {
		reviewersIDs = append(reviewersIDs, reviewers[i].ID)
	}

	now := time.Now()
	pullRequest := &prDomain.PullRequest{
		PullRequestShort: prDomain.PullRequestShort{
			ID:       prID,
			Name:     prName,
			AuthorID: author.ID,
			Status:   prDomain.StatusOpen,
		},
		AssignedReviewers: reviewersIDs,
		CreatedAt:         &now,
	}

	_, err = s.pullRequestRepo.Create(ctx, pullRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pullRequest, nil
}
