package integration_tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

type TeamGetSuite struct {
	BaseSuite
}

func (s *TeamGetSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *TeamGetSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TeamGetSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/team_get"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestTeamGetSuite_Run(t *testing.T) {
	suite.Run(t, new(TeamGetSuite))
}

func (s *TeamGetSuite) TestGetTeamNotFound() {
	const teamName = "not_found_team"

	res, err := s.server.Client().Get(s.server.URL + "/team/get?team_name=" + teamName)
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusNotFound, res.StatusCode)

	var response handlers.ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found"
  }
}
`

	JSONEq(s.T(), expected, response)
}

func (s *TeamGetSuite) TestGetTeamDefault() {
	const teamName = "payments"

	res, err := s.server.Client().Get(s.server.URL + "/team/get?team_name=" + teamName)
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	var response teamsHandler.GetTeamResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Equal(teamName, response.TeamName)
	s.Require().Len(response.Members, 3)

	expectedMembers := map[string]teamsHandler.MemberResponse{
		"u1_Alice": {
			ID:       "u1_Alice",
			Name:     "Alice",
			IsActive: true,
		},
		"u2_Bob": {
			ID:       "u2_Bob",
			Name:     "Bob",
			IsActive: false,
		},
		"u3_John": {
			ID:       "u3_John",
			Name:     "John",
			IsActive: true,
		},
	}

	for _, member := range response.Members {
		expected, exists := expectedMembers[member.ID]
		s.Require().True(exists, "Unexpected user ID: %s", member.ID)
		s.Require().Equal(expected.Name, member.Name)
		s.Require().Equal(expected.IsActive, member.IsActive)
	}
}

func (s *TeamGetSuite) TestGetTeamValidation() {
	const realTeamName = "payments"

	res, err := s.server.Client().Get(s.server.URL + "/team/get?team_name=" + realTeamName)
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	const emptyTeamName = ""

	res, err = s.server.Client().Get(s.server.URL + "/team/get?team_name=" + emptyTeamName)
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusBadRequest, res.StatusCode)

	// no query param
	res, err = s.server.Client().Get(s.server.URL + "/team/get")
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusBadRequest, res.StatusCode)

	// wrong query param
	res, err = s.server.Client().Get(s.server.URL + "/team/get?asdf=123")
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusBadRequest, res.StatusCode)
}
