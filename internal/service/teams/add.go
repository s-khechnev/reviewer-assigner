package teams

import (
	"context"
	"fmt"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
)

func (s *TeamService) AddTeam(
	ctx context.Context,
	name string,
	members []teamsDomain.Member,
) (*teamsDomain.Team, error) {
	team, err := s.teamRepo.GetTeam(ctx, name)
	if err == nil {
		return s.UpdateExistingTeam(ctx, team, members)
	}
	s.log.Error("failed to get team", logger.ErrAttr(err))

	createdTeam, err := s.teamRepo.CreateTeam(ctx, name, members)
	if err != nil {
		s.log.Error("failed to create team", logger.ErrAttr(err))
		return nil, fmt.Errorf("failed to add team: %w", err)
	}

	return createdTeam, nil
}

// UpdateExistingTeam updates the members of an existing team only if the member IDs remain the same
func (s *TeamService) UpdateExistingTeam(
	ctx context.Context,
	oldTeam *teamsDomain.Team,
	newMembers []teamsDomain.Member,
) (*teamsDomain.Team, error) {
	if !areMemberIdsEqual(oldTeam.Members, newMembers) {
		s.log.Info("members are not equal")
		return nil, fmt.Errorf("failed to update existing team: %w", service.ErrTeamAlreadyExists)
	}

	updatedTeam, err := s.teamRepo.UpdateTeam(ctx, oldTeam.Name, newMembers)
	if err != nil {
		s.log.Error("failed to update existing team", logger.ErrAttr(err))
		return nil, fmt.Errorf("failed to update existing team: %w", err)
	}

	return updatedTeam, nil
}

func areMemberIdsEqual(oldMembers, newMembers []teamsDomain.Member) bool {
	if len(oldMembers) != len(newMembers) {
		return false
	}

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
