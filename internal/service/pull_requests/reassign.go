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
		if member.Id != oldReviewer.Id && member.Id != pullRequest.AuthorId {
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
			pullRequest.AssignedReviewers[i] = newReviewer.Id
			break
		}
	}

	updatedPullRequest, err := s.pullRequestRepo.Update(ctx, pullRequestID, pullRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to update pull requesr: %w", err)
	}

	return updatedPullRequest, nil
}
