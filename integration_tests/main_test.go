package integration_tests

//
//import (
//	"bytes"
//	"context"
//	"database/sql"
//	"encoding/json"
//	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
//	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
//	"github.com/gin-gonic/gin"
//	"github.com/go-testfixtures/testfixtures/v3"
//	"github.com/pressly/goose/v3"
//	"github.com/stretchr/testify/suite"
//	"io"
//	"log/slog"
//	"net/http"
//	"net/http/httptest"
//	"reviewer-assigner/internal/app"
//	reviewerPicker "reviewer-assigner/internal/domain/pull_requests/reviewer_pickers"
//	reviewerAssigner "reviewer-assigner/internal/domain/pull_requests/reviewer_reassigners"
//	prHandler "reviewer-assigner/internal/http/handlers/pull_requests"
//	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
//	usersHandler "reviewer-assigner/internal/http/handlers/users"
//	prService "reviewer-assigner/internal/service/pull_requests"
//	teamsService "reviewer-assigner/internal/service/teams"
//	usersService "reviewer-assigner/internal/service/users"
//	"reviewer-assigner/internal/storage/postgres"
//	pullRequestsRepo "reviewer-assigner/internal/storage/pull_requests"
//	teamsRepo "reviewer-assigner/internal/storage/teams"
//	usersRepo "reviewer-assigner/internal/storage/users"
//	"testing"
//	"time"
//)
//
//const migrationsPath = "../migrations/postgres"
//
//var out = io.Discard
//
//// var out = os.Stdout
//
//type TestSuite struct {
//	suite.Suite
//	psqlContainer *PostgreSQLContainer
//	server        *httptest.Server
//	loader        *FixtureLoader
//}
//
//func (s *TestSuite) SetupSuite() {
//	l := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
//	gin.DefaultWriter = out
//
//	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer ctxCancel()
//
//	psqlContainer, err := NewPostgreSQLContainer(ctx)
//	s.Require().NoError(err)
//	s.psqlContainer = psqlContainer
//
//	db, err := sql.Open("postgres", psqlContainer.GetDSN())
//	s.Require().NoError(err)
//
//	err = goose.Up(db, migrationsPath)
//	s.Require().NoError(err)
//
//	err = db.Close()
//	s.Require().NoError(err)
//
//	pool, err := postgres.NewPool(ctx, psqlContainer.GetDSN())
//	s.Require().NoError(err)
//
//	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))
//
//	teamRepo := teamsRepo.NewPostgresTeamRepository(pool, trmpgx.DefaultCtxGetter)
//	userRepo := usersRepo.NewPostgresUserRepository(pool, trmpgx.DefaultCtxGetter)
//	pullRequestRepo := pullRequestsRepo.NewPostgresPullRequestRepository(pool, trmpgx.DefaultCtxGetter)
//
//	teamService := teamsService.NewTeamService(l, teamRepo, txManager)
//	userService := usersService.NewUserService(l, userRepo, pullRequestRepo, txManager)
//	pullRequestService := prService.NewPullRequestService(
//		l,
//		userRepo,
//		teamRepo,
//		pullRequestRepo,
//		&reviewerPicker.RandomReviewerPicker{},
//		reviewerAssigner.NewRandomReviewerReassigner(),
//		txManager,
//	)
//
//	teamHandler := teamsHandler.NewTeamHandler(l, teamService)
//	userHandler := usersHandler.NewUserHandler(l, userService)
//	pullRequestHandler := prHandler.NewPullRequestHandler(l, pullRequestService)
//
//	s.server = httptest.NewServer(app.NewRouter(l, teamHandler, userHandler, pullRequestHandler))
//
//	s.loader = NewFixtureLoader(s.T(), Fixtures)
//}
//
//func (s *TestSuite) TearDownSuite() {
//	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer ctxCancel()
//
//	s.Require().NoError(s.psqlContainer.Terminate(ctx))
//
//	s.server.Close()
//}
//
//func (s *TestSuite) SetupTest() {
//	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
//	s.Require().NoError(err)
//
//	fixtures, err := testfixtures.New(
//		testfixtures.Database(db),
//		testfixtures.Dialect("postgres"),
//	)
//	s.Require().NoError(err)
//	s.Require().NoError(fixtures.Load())
//
//	err = goose.Reset(db, migrationsPath)
//	s.Require().NoError(err)
//
//	err = goose.Up(db, migrationsPath)
//	s.Require().NoError(err)
//}
//
//func TestSuite_Run(t *testing.T) {
//	suite.Run(t, new(TestSuite))
//}
//
//func (s *TestSuite) TestCreateTeam() {
//	requestBody := s.loader.LoadString("fixtures/api/create_team_request.json")
//
//	res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
//	s.Require().NoError(err)
//
//	defer res.Body.Close()
//
//	s.Require().Equal(http.StatusCreated, res.StatusCode)
//
//	response := teamsHandler.AddTeamResponse{}
//	err = json.NewDecoder(res.Body).Decode(&response)
//	s.Require().NoError(err)
//
//	expected := s.loader.LoadString("fixtures/api/create_team_response.json")
//
//	JSONEq(s.T(), expected, response)
//}
//
//func (s *TestSuite) TestGetTeam() {
//	res, err := s.server.Client().Get(s.server.URL + "/team/get?team_name=payments")
//	s.Require().NoError(err)
//
//	defer res.Body.Close()
//
//	s.Require().Equal(http.StatusOK, res.StatusCode)
//
//	response := teamsHandler.GetTeamResponse{}
//	err = json.NewDecoder(res.Body).Decode(&response)
//	s.Require().NoError(err)
//
//	expected := s.loader.LoadString("fixtures/api/get_team_response.json")
//
//	JSONEq(s.T(), expected, response)
//}
