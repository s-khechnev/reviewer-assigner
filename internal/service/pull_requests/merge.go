package pull_requests

import (
	"context"
	"fmt"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
)

// TODO: errors

func (s *PullRequestService) Merge(ctx context.Context, pullRequestId string) (*prDomain.PullRequest, error) {
	mergedPR, err := s.pullRequestRepo.Merge(ctx, pullRequestId)
	if err != nil {
		return nil, fmt.Errorf("failed to merge pull request: %w", err)
	}

	return mergedPR, nil
}
