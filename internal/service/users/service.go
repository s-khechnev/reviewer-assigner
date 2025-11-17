package users

import (
	"context"
	"log/slog"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	usersDomain "reviewer-assigner/internal/domain/users"
)

type UserRepository interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*usersDomain.User, error)
}

type PullRequestRepository interface {
	GetPullRequestsForReview(ctx context.Context, userID string) ([]prDomain.PullRequestShort, error)
}

type UserService struct {
	userRepo UserRepository
	prRepo   PullRequestRepository

	log *slog.Logger
}

func NewUserService(log *slog.Logger, userRepo UserRepository, prRepo PullRequestRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
		log:      log,
	}
}
