package teams

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reviewer-assigner/internal/domain"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
)

func (s *TeamService) AddTeam(
	ctx context.Context,
	name string,
	members []teamsDomain.Member,
) (team *teamsDomain.Team, err error) {
	const op = "services.teams.AddTeam"
	log := s.log.With(
		slog.String("op", op),
		slog.String("team_name", name),
	)

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		team, err = s.teamRepo.GetTeamByName(ctx, name)
		if err != nil {
			log.Warn("team not found")

			team, err = s.createTeam(ctx, name, members)
			return err
		}

		team, err = s.updateExistingTeam(ctx, team, members)

		return err
	})

	return team, err
}

func (s *TeamService) createTeam(
	ctx context.Context,
	teamName string,
	members []teamsDomain.Member,
) (*teamsDomain.Team, error) {
	const op = "services.teams.createTeam"
	log := s.log.With(
		slog.String("op", op),
		slog.String("team_name", teamName),
	)

	_, err := s.teamRepo.SaveTeam(ctx, teamName, members)
	if err != nil {
		log.Error("failed to save team", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to save team: %w", err)
	}

	log.Info("new team saved")

	return &teamsDomain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

// UpdateExistingTeam updates the members of an existing team only if the member IDs remain the same.
func (s *TeamService) updateExistingTeam(
	ctx context.Context,
	team *teamsDomain.Team,
	newMembers []teamsDomain.Member,
) (*teamsDomain.Team, error) {
	const op = "services.teams.updateExistingTeam"
	log := s.log.With(
		slog.String("op", op),
		slog.String("team_name", team.Name),
	)

	err := team.UpdateMembers(newMembers)
	if errors.Is(err, domain.ErrTeamMembersMismatch) {
		log.Warn("members mismatch",
			slog.Any("oldMembers", team.Members),
			slog.Any("newMembers", newMembers),
		)

		return nil, service.ErrTeamAlreadyExists
	}
	if err != nil {
		log.Error("failed to update members", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to update members: %w", err)
	}

	log.Info("team members updated", slog.Any("team", team))

	err = s.teamRepo.UpdateMembers(ctx, team.Name, team.Members)
	if err != nil {
		log.Error("failed to update existing team", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to update existing team: %w", err)
	}

	log.Info("updated team members saved")

	return team, nil
}
