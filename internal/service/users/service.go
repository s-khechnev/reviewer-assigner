package users

import (
	"context"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"log/slog"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	usersDomain "reviewer-assigner/internal/domain/users"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (*usersDomain.User, error)
	UpdateIsActive(ctx context.Context, user *usersDomain.User) error
}

type PullRequestRepository interface {
	GetPullRequestsForReview(ctx context.Context, userID string) ([]prDomain.PullRequestShort, error)
}

type UserService struct {
	userRepo UserRepository
	prRepo   PullRequestRepository

	txManager trm.Manager

	log *slog.Logger
}

func NewUserService(
	log *slog.Logger,
	userRepo UserRepository,
	prRepo PullRequestRepository,
	txManager trm.Manager,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		prRepo:    prRepo,
		log:       log,
		txManager: txManager,
	}
}
