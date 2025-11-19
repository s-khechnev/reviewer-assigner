package pull_requests

import (
	"context"
	"fmt"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/service"
)

func (s *PullRequestService) Reassign(ctx context.Context, pullRequestID, oldReviewerID string) (*prDomain.PullRequest, error) {
	pullRequest, err := s.pullRequestRepo.Get(ctx, pullRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	oldReviewer, err := s.userRepo.Get(ctx, oldReviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get old reviewer: %w", service.ErrUserNotFound)
	}

	activeMembers, err := s.teamRepo.GetActiveMembers(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get active members: %w", err)
	}

	activeMembersExcludeOldReviewerAuthor := make([]teamsDomain.Member, 0, len(activeMembers))
	for _, member := range activeMembers {
		if member.ID != oldReviewer.ID && member.ID != pullRequest.AuthorID {
			activeMembersExcludeOldReviewerAuthor = append(activeMembersExcludeOldReviewerAuthor, member)
		}
	}

	newReviewers, err := s.reviewerReassigner.Pick(activeMembersExcludeOldReviewerAuthor)
	if err != nil {
		return nil, fmt.Errorf("failed to pick reviewers: %w", err)
	}

	if len(newReviewers) != 1 {
		panic(fmt.Sprintf("expected 1 reviewer, got %d", len(newReviewers)))
	}
	newReviewer := newReviewers[0]

	for i, reviewerID := range pullRequest.AssignedReviewers {
		if reviewerID == oldReviewerID {
			pullRequest.AssignedReviewers[i] = newReviewer.ID
			break
		}
	}

	err = s.pullRequestRepo.UpdateReviewers(ctx, pullRequestID, pullRequest.AssignedReviewers)
	if err != nil {
		return nil, fmt.Errorf("failed to update reviewers: %w", err)
	}

	return pullRequest, nil
}
