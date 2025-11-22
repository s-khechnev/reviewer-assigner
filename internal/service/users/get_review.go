package users

import (
	"context"
	"fmt"
	"log/slog"
	"reviewer-assigner/internal/logger"

	prDomain "reviewer-assigner/internal/domain/pullrequests"
)

func (s *UserService) GetReview(
	ctx context.Context,
	userID string,
) ([]prDomain.PullRequestShort, error) {
	const op = "services.users.GetReview"
	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
	)

	prsForReview, err := s.prRepo.GetPullRequestsForReview(ctx, userID)
	if err != nil {
		log.Error("failed to get pull requests for review", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to get pull requests for review: %w", err)
	}

	log.Info("got pull requests for review", slog.Any("pull_requests", prsForReview))

	return prsForReview, nil
}
