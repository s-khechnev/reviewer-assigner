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
	"slices"
)

func (s *PullRequestService) Reassign(ctx context.Context, pullRequestID, oldReviewerID string) (*prDomain.PullRequest, error) {
	const op = "services.pull_requests.Reassign"
	log := s.log.With(
		slog.String("op", op),
		slog.String("pull_request_id", pullRequestID),
		slog.String("old_reviewer_id", oldReviewerID),
	)

	pullRequest, err := s.pullRequestRepo.Get(ctx, pullRequestID)
	if errors.Is(err, service.ErrPullRequestNotFound) {
		log.Error("pull request not found")

		return nil, service.ErrPullRequestNotFound
	}
	if err != nil {
		log.Error("failed to get pull request", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	log.Info("got pull request", slog.Any("pull_request", pullRequest))

	if pullRequest.Status == prDomain.StatusMerged {
		log.Info("pull request is already merged")

		return nil, service.ErrPullRequestAlreadyMerged
	}

	oldReviewer, err := s.userRepo.Get(ctx, oldReviewerID)
	if errors.Is(err, service.ErrUserNotFound) {
		log.Error("old reviewer not found")

		return nil, service.ErrPullRequestNotFound
	}
	if err != nil {
		log.Error("failed to get old reviewer", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to get old reviewer: %w", err)
	}

	log.Info("got old reviewer", slog.Any("old_reviewer", oldReviewer))

	if idx := slices.Index(pullRequest.AssignedReviewers, oldReviewer.ID); idx == -1 {
		log.Error("old reviewer is not assigned to this PR")

		return nil, service.ErrPullRequestNotAssigned
	}

	team, err := s.teamRepo.GetTeam(ctx, oldReviewer.TeamName)
	if errors.Is(err, service.ErrTeamNotFound) {
		log.Error("team not found")

		return nil, service.ErrTeamNotFound
	}
	if err != nil {
		log.Error("failed to get team", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	log.Info("got team", slog.Any("team", team))

	isAlreadyReviewer := func(member *teamsDomain.Member) bool {
		return slices.Index(pullRequest.AssignedReviewers, member.ID) != -1
	}

	const activeMembersDefaultCap = 4
	activeMembersExcludeOldReviewerAuthor := make([]teamsDomain.Member, 0, activeMembersDefaultCap)
	for _, member := range team.Members {
		if member.IsActive &&
			member.ID != oldReviewer.ID &&
			member.ID != pullRequest.AuthorID &&
			!isAlreadyReviewer(&member) {
			activeMembersExcludeOldReviewerAuthor = append(activeMembersExcludeOldReviewerAuthor, member)
		}
	}

	newReviewer, err := s.reviewerReassigner.Reassign(&oldReviewer.Member, activeMembersExcludeOldReviewerAuthor)
	if errors.Is(err, prDomain.ErrNotEnoughMembers) {
		log.Error("not enough active members")

		return nil, fmt.Errorf("failed to reassign: %w", service.ErrPullRequestNoCandidates)
	}
	if err != nil {
		log.Error("failed to reassign", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to reassign: %w", err)
	}

	log.Info("got new reviewer", slog.String("reviewer_id", newReviewer.ID))

	for i, reviewerID := range pullRequest.AssignedReviewers {
		if reviewerID == oldReviewerID {
			pullRequest.AssignedReviewers[i] = newReviewer.ID
			break
		}
	}

	err = s.pullRequestRepo.UpdateReviewers(ctx, pullRequestID, pullRequest.AssignedReviewers)
	if err != nil {
		log.Error("failed to update reviewers", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to update reviewers: %w", err)
	}

	return pullRequest, nil
}
