package pull_requests

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	prDomain "reviewer-assigner/internal/domain/pull_requests"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"time"
)

// TODO: transactions, errors

func (s *PullRequestService) Create(ctx context.Context, prID, prName, authorID string) (*prDomain.PullRequest, error) {
	const op = "services.pull_requests.Create"
	log := s.log.With(
		slog.String("op", op),
		slog.String("pull_request_id", prID),
		slog.String("pull_request_name", prName),
		slog.String("author_id", authorID),
	)

	author, err := s.userRepo.Get(ctx, authorID)
	if errors.Is(err, service.ErrUserNotFound) {
		log.Error("failed to find author")

		return nil, service.ErrUserNotFound
	}
	if err != nil {
		log.Error("failed to get author", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	log.Info("got author", slog.Any("author", author))

	log = log.With(slog.String("team_name", author.TeamName))

	team, err := s.teamRepo.GetTeam(ctx, author.TeamName)
	if errors.Is(err, service.ErrTeamNotFound) {
		log.Error("team not found")

		return nil, service.ErrTeamNotFound
	}
	if err != nil {
		log.Error("failed to find team")

		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	log.Info("got team", slog.Any("team", team))

	const activeMembersDefaultCap = 4
	activeMembersExcludeAuthor := make([]teamsDomain.Member, 0, activeMembersDefaultCap)
	for _, member := range team.Members {
		if member.IsActive && member.ID != author.ID {
			activeMembersExcludeAuthor = append(activeMembersExcludeAuthor, member)
		}
	}

	log.Info("got active members", slog.Any("activeMembers", activeMembersExcludeAuthor))

	const countReviewers = 2
	reviewers := s.reviewerPicker.Pick(activeMembersExcludeAuthor, countReviewers)

	log.Info("got reviewers", slog.Any("reviewers", reviewers))

	reviewerIDs := make([]string, 0, len(reviewers))
	for _, reviewer := range reviewers {
		reviewerIDs = append(reviewerIDs, reviewer.ID)
	}

	now := time.Now()
	pullRequest := &prDomain.PullRequest{
		PullRequestShort: prDomain.PullRequestShort{
			ID:       prID,
			Name:     prName,
			AuthorID: author.ID,
			Status:   prDomain.StatusOpen,
		},
		AssignedReviewers: reviewerIDs,
		CreatedAt:         &now,
	}

	id, err := s.pullRequestRepo.Create(ctx, pullRequest)
	if errors.Is(err, service.ErrPullRequestAlreadyExists) {
		log.Error("pull request already exists")

		return nil, service.ErrPullRequestAlreadyExists
	}
	if err != nil {
		log.Error("failed to create pull request", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	log.Info("pull request created", slog.String("id", id), slog.Any("pull_request", pullRequest))

	return pullRequest, nil
}
