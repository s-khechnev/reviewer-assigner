package teams

import (
	"context"
	"fmt"
	"log/slog"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
)

func (s *TeamService) AddTeam(
	ctx context.Context,
	name string,
	members []teamsDomain.Member,
) (*teamsDomain.Team, error) {
	const op = "services.teams.AddTeam"
	log := s.log.With(
		slog.String("op", op),
		slog.String("team_name", name),
	)

	// if team already exists then update members
	if team, err := s.teamRepo.GetTeam(ctx, name); err == nil {
		log.Info("team already exists")

		return s.UpdateExistingTeam(ctx, team, members)
	}

	log.Info("failed to find team")

	id, err := s.teamRepo.SaveTeam(ctx, name, members)
	if err != nil {
		log.Error("failed to save team", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to add team: %w", err)
	}

	log.Info("added new team", slog.Int64("id", id))

	return &teamsDomain.Team{
		Name:    name,
		Members: members,
	}, nil
}

// UpdateExistingTeam updates the members of an existing team only if the member IDs remain the same
func (s *TeamService) UpdateExistingTeam(
	ctx context.Context,
	oldTeam *teamsDomain.Team,
	newMembers []teamsDomain.Member,
) (*teamsDomain.Team, error) {
	const op = "services.teams.UpdateExistingTeam"
	log := s.log.With(
		slog.String("op", op),
		slog.String("team_name", oldTeam.Name),
	)

	if !areMemberIDsEqual(oldTeam.Members, newMembers) {
		log.Warn("member IDs mismatch",
			slog.Any("oldMembers", oldTeam.Members),
			slog.Any("newMembers", newMembers),
		)

		return nil, fmt.Errorf("member IDs mismatch: %w", service.ErrTeamAlreadyExists)
	}

	err := s.teamRepo.UpdateTeam(ctx, oldTeam.Name, newMembers)
	if err != nil {
		log.Error("failed to update existing team", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to update existing team: %w", err)
	}

	log.Info("team updated", slog.Any("members", newMembers))

	return &teamsDomain.Team{
		Name:    oldTeam.Name,
		Members: newMembers,
	}, nil
}

func areMemberIDsEqual(oldMembers, newMembers []teamsDomain.Member) bool {
	set := make(map[string]struct{}, len(oldMembers))
	for _, m := range oldMembers {
		set[m.ID] = struct{}{}
	}

	for _, m := range newMembers {
		if _, ok := set[m.ID]; !ok {
			return false
		}
	}

	return true
}
