package pullrequests

import (
	"context"
	"log/slog"
	prsDomain "reviewer-assigner/internal/domain/pullrequests"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	usersDomain "reviewer-assigner/internal/domain/users"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (*usersDomain.User, error)
}

type TeamRepository interface {
	GetTeamByName(ctx context.Context, teamName string) (*teamsDomain.Team, error)
}

type PullRequestRepository interface {
	GetByID(ctx context.Context, pullRequestID string) (*prsDomain.PullRequest, error)
	Create(ctx context.Context, pullRequest *prsDomain.PullRequest) (string, error)
	SetStatusMerged(ctx context.Context, pullRequestID string, mergedAt time.Time) error
	UpdateReviewers(ctx context.Context, pullRequestID string, newReviewerIDs []string) error
}

type ReviewerPicker interface {
	Pick(members []teamsDomain.Member, count int) []teamsDomain.Member
}

type ReviewerReassigner interface {
	Reassign(
		oldReviewer *teamsDomain.Member,
		members []teamsDomain.Member,
	) (newReviewer *teamsDomain.Member, err error)
}

type PullRequestService struct {
	userRepo        UserRepository
	teamRepo        TeamRepository
	pullRequestRepo PullRequestRepository

	reviewerPicker     ReviewerPicker
	reviewerReassigner ReviewerReassigner

	txManager trm.Manager

	log *slog.Logger
}

func NewPullRequestService(
	log *slog.Logger,
	userRepo UserRepository,
	teamRepo TeamRepository,
	pullRequestRepo PullRequestRepository,
	reviewerPicker ReviewerPicker,
	reviewerReassigner ReviewerReassigner,
	txManager trm.Manager,
) *PullRequestService {
	return &PullRequestService{
		userRepo:        userRepo,
		teamRepo:        teamRepo,
		pullRequestRepo: pullRequestRepo,

		reviewerPicker:     reviewerPicker,
		reviewerReassigner: reviewerReassigner,

		txManager: txManager,

		log: log,
	}
}
