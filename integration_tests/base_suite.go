package integration_tests

import (
	"context"
	"database/sql"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"io"
	"log/slog"
	"net/http/httptest"
	"reviewer-assigner/internal/app"
	reviewerPicker "reviewer-assigner/internal/domain/pull_requests/reviewer_pickers"
	reviewerAssigner "reviewer-assigner/internal/domain/pull_requests/reviewer_reassigners"
	prHandler "reviewer-assigner/internal/http/handlers/pull_requests"
	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
	usersHandler "reviewer-assigner/internal/http/handlers/users"
	prService "reviewer-assigner/internal/service/pull_requests"
	teamsService "reviewer-assigner/internal/service/teams"
	usersService "reviewer-assigner/internal/service/users"
	"reviewer-assigner/internal/storage/postgres"
	pullRequestsRepo "reviewer-assigner/internal/storage/pull_requests"
	teamsRepo "reviewer-assigner/internal/storage/teams"
	usersRepo "reviewer-assigner/internal/storage/users"
	"time"
)

const migrationsPath = "../migrations/postgres"

var out = io.Discard

// var out = os.Stdout

type BaseSuite struct {
	suite.Suite
	psqlContainer *PostgreSQLContainer
	server        *httptest.Server
	loader        *FixtureLoader
}

func (s *BaseSuite) SetupSuite() {
	l := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
	gin.DefaultWriter = out

	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	psqlContainer, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	db, err := sql.Open("postgres", psqlContainer.GetDSN())
	s.Require().NoError(err)

	err = goose.Up(db, migrationsPath)
	s.Require().NoError(err)

	err = db.Close()
	s.Require().NoError(err)

	pool, err := postgres.NewPool(ctx, psqlContainer.GetDSN())
	s.Require().NoError(err)

	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	teamRepo := teamsRepo.NewPostgresTeamRepository(pool, trmpgx.DefaultCtxGetter)
	userRepo := usersRepo.NewPostgresUserRepository(pool, trmpgx.DefaultCtxGetter)
	pullRequestRepo := pullRequestsRepo.NewPostgresPullRequestRepository(pool, trmpgx.DefaultCtxGetter)

	teamService := teamsService.NewTeamService(l, teamRepo, txManager)
	userService := usersService.NewUserService(l, userRepo, pullRequestRepo, txManager)
	pullRequestService := prService.NewPullRequestService(
		l,
		userRepo,
		teamRepo,
		pullRequestRepo,
		&reviewerPicker.RandomReviewerPicker{},
		reviewerAssigner.NewRandomReviewerReassigner(),
		txManager,
	)

	teamHandler := teamsHandler.NewTeamHandler(l, teamService)
	userHandler := usersHandler.NewUserHandler(l, userService)
	pullRequestHandler := prHandler.NewPullRequestHandler(l, pullRequestService)

	s.server = httptest.NewServer(app.NewRouter(l, teamHandler, userHandler, pullRequestHandler))

	s.loader = NewFixtureLoader(s.T(), Fixtures)
}

func (s *BaseSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.psqlContainer.Terminate(ctx))

	s.server.Close()
}
