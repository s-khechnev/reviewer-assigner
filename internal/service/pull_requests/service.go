package pull_requests

import (
	"context"
	"log/slog"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	"reviewer-assigner/internal/domain/pull_requests/reviewer_pickers"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	usersDomain "reviewer-assigner/internal/domain/users"
)

type UserRepository interface {
	Get(ctx context.Context, userID string) (*usersDomain.User, error)
}

type TeamRepository interface {
	GetActiveMembers(ctx context.Context, teamName string) ([]teamsDomain.Member, error)
}

type PullRequestRepository interface {
	Get(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error)
	Create(ctx context.Context, pullRequest *prDomain.PullRequest) (string, error)
	Merge(ctx context.Context, pullRequestID string) (*prDomain.PullRequest, error)
	UpdateReviewers(ctx context.Context, pullRequestID string, newReviewerIDs []string) error
}

type ReviewerPicker interface {
	Pick(members []teamsDomain.Member) ([]teamsDomain.Member, error)
}

type PullRequestService struct {
	userRepo        UserRepository
	teamRepo        TeamRepository
	pullRequestRepo PullRequestRepository

	reviewerPicker     ReviewerPicker
	reviewerReassigner *reviewer_pickers.RandomReviewerPicker

	log *slog.Logger
}

func NewPullRequestService(log *slog.Logger,
	userRepo UserRepository,
	teamRepo TeamRepository,
	pullRequestRepo PullRequestRepository,
	reviewerPicker ReviewerPicker,
) *PullRequestService {
	return &PullRequestService{
		userRepo:        userRepo,
		teamRepo:        teamRepo,
		pullRequestRepo: pullRequestRepo,

		reviewerPicker:     reviewerPicker,
		reviewerReassigner: reviewer_pickers.NewRandomReviewerPicker(1),

		log: log,
	}
}
