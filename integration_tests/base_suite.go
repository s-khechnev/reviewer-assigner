package integration_tests

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http/httptest"
	"reviewer-assigner/internal/app"
	reviewerPicker "reviewer-assigner/internal/domain/pullrequests/pickers"
	reviewerAssigner "reviewer-assigner/internal/domain/pullrequests/reassigners"
	prsHandler "reviewer-assigner/internal/http/handlers/pullrequests"
	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
	usersHandler "reviewer-assigner/internal/http/handlers/users"
	prsService "reviewer-assigner/internal/service/pullrequests"
	teamsService "reviewer-assigner/internal/service/teams"
	usersService "reviewer-assigner/internal/service/users"
	"reviewer-assigner/internal/storage/postgres"
	prsRepo "reviewer-assigner/internal/storage/pullrequests"
	teamsRepo "reviewer-assigner/internal/storage/teams"
	usersRepo "reviewer-assigner/internal/storage/users"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
)

const migrationsPath = "../migrations/postgres"

//gochecknoglobals:ignore
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
	gin.DefaultWriter = out // nolint:reassign

	const baseTimout = 30 * time.Second
	ctx, ctxCancel := context.WithTimeout(context.Background(), baseTimout)
	defer ctxCancel()

	psqlContainer, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	db, err := sql.Open("postgres", psqlContainer.GetDSN())
	s.Require().NoError(err)

	err = goose.SetDialect("postgres")
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
	pullRequestRepo := prsRepo.NewPostgresPullRequestRepository(
		pool,
		trmpgx.DefaultCtxGetter,
	)

	teamService := teamsService.NewTeamService(l, teamRepo, txManager)
	userService := usersService.NewUserService(l, userRepo, pullRequestRepo, txManager)
	pullRequestService := prsService.NewPullRequestService(
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
	pullRequestHandler := prsHandler.NewPullRequestHandler(l, pullRequestService)

	s.server = httptest.NewServer(app.NewRouter(l, teamHandler, userHandler, pullRequestHandler))

	s.loader = NewFixtureLoader(s.T(), Fixtures)
}

func (s *BaseSuite) TearDownSuite() {
	const timeoutToDown = 5 * time.Second
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeoutToDown)
	defer ctxCancel()

	s.Require().NoError(s.psqlContainer.Terminate(ctx))

	s.server.Close()
}

func (s *BaseSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	goose.SetLogger(goose.NopLogger())
	err = goose.Reset(db, migrationsPath)
	s.Require().NoError(err)

	err = goose.Up(db, migrationsPath)
	s.Require().NoError(err)
}
