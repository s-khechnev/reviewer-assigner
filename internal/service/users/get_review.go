package users

import (
	"context"
	"fmt"

	prDomain "reviewer-assigner/internal/domain/pull_requests"
)

func (s *UserService) GetReview(ctx context.Context, userID string) ([]prDomain.PullRequestShort, error) {
	prsForReview, err := s.prRepo.GetPullRequestsForReview(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests for review: %w", err)
	}

	return prsForReview, nil
}
