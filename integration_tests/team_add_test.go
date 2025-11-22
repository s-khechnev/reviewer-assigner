package integration_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"net/http"
	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
	"testing"
)

type TeamAddSuite struct {
	BaseSuite
}

func (s *TeamAddSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *TeamAddSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TeamAddSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	err = goose.Reset(db, migrationsPath)
	s.Require().NoError(err)

	err = goose.Up(db, migrationsPath)
	s.Require().NoError(err)
}

func TestTeamAddSuite_Run(t *testing.T) {
	suite.Run(t, new(TeamAddSuite))
}

func (s *TeamAddSuite) TestAddEmptyTeam() {
	requestBody := `
	{
  		"team_name": "payments",
  		"members": []
	}
`

	res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
	{
	  "team": {
		"team_name": "payments",
		"members": []
      }
	}
`

	JSONEq(s.T(), expected, response)
}

func (s *TeamAddSuite) TestAddTeam() {
	requestBody := s.loader.LoadString("fixtures/api/add_team_request.json")

	res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := s.loader.LoadString("fixtures/api/add_team_response.json")

	JSONEq(s.T(), expected, response)
}
