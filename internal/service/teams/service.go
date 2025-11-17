package teams

import (
	"context"
	"log/slog"
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type UserRepository interface {
}

type TeamRepository interface {
	GetTeam(ctx context.Context, name string) (*teamsDomain.Team, error)
	CreateTeam(ctx context.Context, name string, members []teamsDomain.Member) (*teamsDomain.Team, error)
	UpdateTeam(ctx context.Context, name string, members []teamsDomain.Member) (*teamsDomain.Team, error)
}

type TeamService struct {
	userRepo UserRepository
	teamRepo TeamRepository

	log *slog.Logger
}

func NewTeamService(log *slog.Logger, userRepo UserRepository, teamRepo TeamRepository) *TeamService {
	return &TeamService{
		userRepo: userRepo,
		teamRepo: teamRepo,
		log:      log,
	}
}
