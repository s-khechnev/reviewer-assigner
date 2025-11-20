package pull_requests

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
)

// TODO: errors

func (s *PullRequestService) Merge(ctx context.Context, pullRequestId string) (*prDomain.PullRequest, error) {
	const op = "services.pull_requests.Merge"
	log := s.log.With(
		slog.String("op", op),
		slog.String("pull_request_id", pullRequestId),
	)

	mergedPR, err := s.pullRequestRepo.Merge(ctx, pullRequestId)
	if errors.Is(err, service.ErrPullRequestNotFound) {
		log.Error("pull request not found")

		return nil, service.ErrPullRequestNotFound
	}
	if err != nil {
		log.Error("failed to merge pull request", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to merge pull request: %w", err)
	}

	log.Info("pull request merged", slog.Any("pull_request", mergedPR))

	return mergedPR, nil
}
