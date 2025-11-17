package teams

import (
	"context"
	"fmt"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/service"
)

func (s *TeamService) GetTeam(
	ctx context.Context,
	name string,
) (*teamsDomain.Team, error) {
	team, err := s.teamRepo.GetTeam(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", service.ErrTeamNotFound)
	}

	return team, nil
}
