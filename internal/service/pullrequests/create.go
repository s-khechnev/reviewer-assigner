package pullrequests

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	prsDomain "reviewer-assigner/internal/domain/pullrequests"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	usersDomain "reviewer-assigner/internal/domain/users"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"time"
)

func (s *PullRequestService) Create(
	ctx context.Context,
	prID, prName, authorID string,
) (pullRequest *prsDomain.PullRequest, err error) {
	const op = "services.pull_requests.Create"
	log := s.log.With(
		slog.String("op", op),
		slog.String("pull_request_id", prID),
		slog.String("pull_request_name", prName),
		slog.String("author_id", authorID),
	)

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		var author *usersDomain.User
		author, err = s.userRepo.GetUserByID(ctx, authorID)
		if errors.Is(err, service.ErrUserNotFound) {
			log.Error("author not found")

			return service.ErrUserNotFound
		}
		if err != nil {
			log.Error("failed to get author", logger.ErrAttr(err))

			return fmt.Errorf("failed to get author: %w", err)
		}

		log.Info("got author", slog.Any("author", author))

		var team *teamsDomain.Team
		team, err = s.teamRepo.GetTeamByName(ctx, author.TeamName)
		if errors.Is(err, service.ErrTeamNotFound) {
			log.Error("team not found")

			return service.ErrTeamNotFound
		}
		if err != nil {
			log.Error("failed to find team")

			return fmt.Errorf("failed to get team: %w", err)
		}

		log.Info("got team", slog.Any("team", team))

		now := time.Now()
		pullRequest = &prsDomain.PullRequest{
			PullRequestShort: prsDomain.PullRequestShort{
				ID:       prID,
				Name:     prName,
				AuthorID: author.ID,
				Status:   prsDomain.StatusOpen,
			},
			CreatedAt: &now,
		}

		const countReviewers = 2
		err = pullRequest.AssignReviewers(team.Members, s.reviewerPicker, countReviewers)
		if err != nil {
			log.Error("failed to assign reviewers", logger.ErrAttr(err))

			return fmt.Errorf("failed to assign reviewers: %w", err)
		}

		log.Info("got reviewers", slog.Any("reviewers", pullRequest.AssignedReviewers))

		_, err = s.pullRequestRepo.Create(ctx, pullRequest)
		if errors.Is(err, service.ErrPullRequestAlreadyExists) {
			log.Error("pull request already exists")

			return service.ErrPullRequestAlreadyExists
		}
		if err != nil {
			log.Error("failed to create pull request", logger.ErrAttr(err))

			return fmt.Errorf("failed to create pull request: %w", err)
		}

		log.Info("pull request created", slog.Any("pull_request", pullRequest))

		return nil
	})

	return pullRequest, err
}
