package teams

import (
	"context"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"log/slog"
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type TeamRepository interface {
	GetTeamByName(ctx context.Context, name string) (*teamsDomain.Team, error)
	SaveTeam(ctx context.Context, name string, members []teamsDomain.Member) (int64, error)
	UpdateMembers(ctx context.Context, name string, newMembers []teamsDomain.Member) error
}

type TeamService struct {
	teamRepo TeamRepository

	txManager trm.Manager

	log *slog.Logger
}

func NewTeamService(log *slog.Logger, teamRepo TeamRepository, txManager trm.Manager) *TeamService {
	return &TeamService{
		teamRepo:  teamRepo,
		txManager: txManager,
		log:       log,
	}
}
