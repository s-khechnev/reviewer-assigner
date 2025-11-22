package pullrequests

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reviewer-assigner/internal/domain"
	prsDomain "reviewer-assigner/internal/domain/pullrequests"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	usersDomain "reviewer-assigner/internal/domain/users"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"slices"
)

func (s *PullRequestService) Reassign(
	ctx context.Context, pullRequestID, oldReviewerID string,
) (pullRequest *prsDomain.PullRequest, replacedBy string, err error) {
	const op = "services.pull_requests.Reassign"
	log := s.log.With(
		slog.String("op", op),
		slog.String("pull_request_id", pullRequestID),
		slog.String("old_reviewer_id", oldReviewerID),
	)

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		pullRequest, err = s.pullRequestRepo.GetByID(ctx, pullRequestID)
		if errors.Is(err, service.ErrPullRequestNotFound) {
			log.Error("pull request not found")

			return service.ErrPullRequestNotFound
		}
		if err != nil {
			log.Error("failed to get pull request", logger.ErrAttr(err))

			return fmt.Errorf("failed to get repo: %w", err)
		}

		log.Info("got pull request", slog.Any("pull_request", pullRequest))

		if pullRequest.Status == prsDomain.StatusMerged {
			log.Info("pull request is already merged")

			return service.ErrPullRequestAlreadyMerged
		}

		var oldReviewer *usersDomain.User
		oldReviewer, err = s.userRepo.GetUserByID(ctx, oldReviewerID)
		if errors.Is(err, service.ErrUserNotFound) {
			log.Error("old reviewer not found")

			return service.ErrPullRequestNotFound
		}
		if err != nil {
			log.Error("failed to get old reviewer", logger.ErrAttr(err))

			return fmt.Errorf("failed to get old reviewer: %w", err)
		}

		log.Info("got old reviewer", slog.Any("old_reviewer", oldReviewer))

		if idx := slices.Index(pullRequest.AssignedReviewers, oldReviewer.ID); idx == -1 {
			log.Error("old reviewer is not assigned to this PR")

			return service.ErrPullRequestNotAssigned
		}

		var team *teamsDomain.Team
		team, err = s.teamRepo.GetTeamByName(ctx, oldReviewer.TeamName)
		if errors.Is(err, service.ErrTeamNotFound) {
			log.Error("team not found")

			return service.ErrTeamNotFound
		}
		if err != nil {
			log.Error("failed to get team", logger.ErrAttr(err))

			return fmt.Errorf("failed to get members: %w", err)
		}

		log.Info("got team", slog.Any("team", team))

		replacedBy, err = pullRequest.Reassign(
			&oldReviewer.Member,
			team.Members,
			s.reviewerReassigner,
		)
		if errors.Is(err, domain.ErrNotEnoughMembers) {
			log.Error("not enough active members")

			return service.ErrPullRequestNoCandidates
		}
		if err != nil {
			log.Error("failed to reassign", logger.ErrAttr(err))

			return fmt.Errorf("failed to reassign: %w", err)
		}

		err = s.pullRequestRepo.UpdateReviewers(ctx, pullRequestID, pullRequest.AssignedReviewers)
		if err != nil {
			log.Error("failed to update reviewers", logger.ErrAttr(err))

			return fmt.Errorf("failed to update reviewers: %w", err)
		}

		return nil
	})

	return pullRequest, replacedBy, err
}
