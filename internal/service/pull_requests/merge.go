package pull_requests

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reviewer-assigner/internal/domain"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
)

func (s *PullRequestService) Merge(ctx context.Context, pullRequestID string) (pullRequest *prDomain.PullRequest, err error) {
	const op = "services.pull_requests.Merge"
	log := s.log.With(
		slog.String("op", op),
		slog.String("pull_request_id", pullRequestID),
	)

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		pullRequest, err = s.pullRequestRepo.GetByID(ctx, pullRequestID)
		if errors.Is(err, service.ErrPullRequestNotFound) {
			log.Error("pull request not found")

			return service.ErrPullRequestNotFound
		}
		if err != nil {
			log.Error("failed to merge pull request", logger.ErrAttr(err))

			return fmt.Errorf("failed to merge pull request: %w", err)
		}

		err = pullRequest.Merge()
		if errors.Is(err, domain.ErrPullRequestAlreadyMerged) {
			// It's ok
			log.Info("pull request is already merged")

			return nil
		}
		if err != nil {
			log.Error("failed to merge pull request", logger.ErrAttr(err))

			return fmt.Errorf("failed to merge pull request: %w", err)
		}

		err = s.pullRequestRepo.SetStatusMerged(ctx, pullRequestID, *pullRequest.MergedAt)
		if err != nil {
			log.Error("failed to set status merged", logger.ErrAttr(err))

			return fmt.Errorf("failed to set status merged: %w", err)
		}

		return nil
	})

	return pullRequest, err
}
