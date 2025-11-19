package teams

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"reviewer-assigner/internal/storage"
)

func (s *TeamService) GetTeam(
	ctx context.Context,
	name string,
) (*teamsDomain.Team, error) {
	const op = "services.teams.GetTeam"
	log := s.log.With(
		slog.String("op", op),
		slog.String("team_name", name),
	)

	team, err := s.teamRepo.GetTeam(ctx, name)
	if errors.Is(err, storage.ErrTeamNotFound) {
		log.Warn("team not found")
		return nil, service.ErrTeamNotFound
	}
	if err != nil {
		log.Error("failed to get team", logger.ErrAttr(err))
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	log.Info("got team", slog.String("team_name", team.Name))

	return team, nil
}
