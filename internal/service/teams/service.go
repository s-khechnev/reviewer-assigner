package teams

import (
	"context"
	"log/slog"
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type TeamRepository interface {
	GetTeam(ctx context.Context, name string) (*teamsDomain.Team, error)
	SaveTeam(ctx context.Context, name string, members []teamsDomain.Member) (int64, error)
	UpdateTeam(ctx context.Context, name string, newMembers []teamsDomain.Member) error
}

type TeamService struct {
	teamRepo TeamRepository

	log *slog.Logger
}

func NewTeamService(log *slog.Logger, teamRepo TeamRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		log:      log,
	}
}
