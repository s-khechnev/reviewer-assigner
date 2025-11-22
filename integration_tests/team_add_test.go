package integration_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
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

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/team_add"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestTeamAddSuite_Run(t *testing.T) {
	suite.Run(t, new(TeamAddSuite))
}

func (s *TeamAddSuite) TestAddTeamEmpty() {
	requestBody := `
{
	"team_name": "payments",
	"members": []
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
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

func (s *TeamAddSuite) TestAddTeamDefault() {
	requestBody := `
{
  "team_name": "payments",
  "members": [
    {
      "user_id": "u1",
      "username": "Alice",
      "is_active": true
    },
    {
      "user_id": "u2",
      "username": "Bob",
      "is_active": true
    }
  ]
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
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
    "members": [
      {
        "user_id": "u1",
        "username": "Alice",
        "is_active": true
      },
      {
        "user_id": "u2",
        "username": "Bob",
        "is_active": true
      }
    ]
  }
}
`

	JSONEq(s.T(), expected, response)
}

func (s *TeamAddSuite) TestAddTeamAlreadyExists() {
	requestBody := `
{
	"team_name": "backend_already_exists", 
	"members": [
		{
			"user_id": "u2",
			"username": "Bob",
			"is_active": true
		}
	]
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusBadRequest, res.StatusCode)

	var response handlers.ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
  "error": {
    "code": "TEAM_EXISTS",
    "message": "backend_already_exists already exists"
  }
}
`

	JSONEq(s.T(), expected, response)
}

func (s *TeamAddSuite) TestAddTeamWithInactiveMembers() {
	requestBody := `
{
	"team_name": "mixed_team",
	"members": [
		{
			"user_id": "u1",
			"username": "ActiveUser",
			"is_active": true
		},
		{
			"user_id": "u2",
			"username": "InactiveUser", 
			"is_active": false
		}
	]
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
	"team": {
		"team_name": "mixed_team",
		"members": [
			{
				"user_id": "u1",
				"username": "ActiveUser",
				"is_active": true
			},
			{
				"user_id": "u2",
				"username": "InactiveUser",
				"is_active": false
			}
		]
	}
}
`

	JSONEq(s.T(), expected, response)
}

func (s *TeamAddSuite) TestAddTeamSingleMember() {
	requestBody := `
{
	"team_name": "solo_team",
	"members": [
		{
			"user_id": "u_solo",
			"username": "SoloPlayer",
			"is_active": true
		}
	]
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
	"team": {
		"team_name": "solo_team",
		"members": [
			{
				"user_id": "u_solo",
				"username": "SoloPlayer",
				"is_active": true
			}
		]
	}
}
`
	JSONEq(s.T(), expected, response)
}

func (s *TeamAddSuite) TestAddTeamInvalidJSON() {
	invalidJSON := `
{
	"team_name": "invalid_team",
	"members": [
		{
			"user_id": "u1",
			"username": "Alice",
			"is_active": true
		}
	]
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(invalidJSON))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *TeamAddSuite) TestAddTeamMissingRequiredFields() {
	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
	}{
		{
			name:         "missing_team_name",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
	"members": [
		{
			"user_id": "u1",
			"username": "Alice",
			"is_active": true
		}
	]
}
`,
		},
		{
			name:         "missing_members",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
	"team_name": "no_members_team"
}
`,
		},
		{
			name:         "missing_user_id_in_member",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
	"team_name": "team_with_invalid_member",
	"members": [
		{
			"username": "Alice",
			"is_active": true
		}
	]
}
`,
		},
		{
			name:         "missing_username_in_member",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
	"team_name": "team_with_invalid_member",
	"members": [
		{
			"user_id": "uid1",
			"is_active": true
		}
	]
}
`,
		},
		{
			name:         "missing_is_active_in_member",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
	"team_name": "team_with_invalid_member",
	"members": [
		{
			"user_id": "uid1",
			"username": "Alice"
		}
	]
}
`,
		},
		{
			name:         "missing_many_in_member",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
	"team_name": "team_with_invalid_member",
	"members": [
		{
			"user_id": "uid1"
		}
	]
}
`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.server.Client().
				Post(s.server.URL+"/team/add", "", bytes.NewBufferString(tc.requestBody))
			s.Require().NoError(err)
			defer res.Body.Close()

			s.Require().Equal(tc.expectedCode, res.StatusCode)
		})
	}
}

func (s *TeamAddSuite) TestAddTeamUpdatesExistingUsers() {
	// обновляем только двух из трёх
	requestBody := `
{
	"team_name": "backend_already_exists",
	"members": [
		{
			"user_id": "u1_Alice",
			"username": "Alice_NewName",
			"is_active": false
		},
		{
			"user_id": "u2_Bob",
			"username": "Bob_NewName2",
			"is_active": true
		}
	]
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Equal("backend_already_exists", response.TeamName)
	s.Require().Len(response.Members, 3)

	expected := map[string]teamsHandler.MemberResponse{
		"u1_Alice": {
			ID:       "u1_Alice",
			Name:     "Alice_NewName",
			IsActive: false,
		},
		"u2_Bob": {
			ID:       "u2_Bob",
			Name:     "Bob_NewName2",
			IsActive: true,
		},
		"u3_John": {
			ID:       "u3_John",
			Name:     "John",
			IsActive: true,
		},
	}

	for _, member := range response.Members {
		expectedMember, ok := expected[member.ID]
		s.Require().True(ok)
		s.Require().Equal(expectedMember.Name, member.Name)
		s.Require().Equal(expectedMember.IsActive, member.IsActive)
	}
}

func (s *TeamAddSuite) TestAddTeamUpdatesAllUsers() {
	requestBody := `
{
	"team_name": "frontend",
	"members": [
		{
			"user_id": "u4_Vanya",
			"username": "newNameVanya",
			"is_active": true
		},
		{
			"user_id": "u5_Petya",
			"username": "newNamePetya",
			"is_active": true
		}
	]
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/team/add", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Equal("frontend", response.TeamName)
	s.Require().Len(response.Members, 2)

	expected := map[string]teamsHandler.MemberResponse{
		"u4_Vanya": {
			ID:       "u4_Vanya",
			Name:     "newNameVanya",
			IsActive: true,
		},
		"u5_Petya": {
			ID:       "u5_Petya",
			Name:     "newNamePetya",
			IsActive: true,
		},
	}

	for _, member := range response.Members {
		expectedMember, ok := expected[member.ID]
		s.Require().True(ok)
		s.Require().Equal(expectedMember.Name, member.Name)
		s.Require().Equal(expectedMember.IsActive, member.IsActive)
	}
}
