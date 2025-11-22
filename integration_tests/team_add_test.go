package integration_tests

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
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
	s.BaseSuite.SetupTest()
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

func (s *TeamAddSuite) TestAddTeamDuplicate() {
	firstRequestBody := `
{
	"team_name": "backend",
	"members": [
		{
			"user_id": "u1",
			"username": "Alice",
			"is_active": true
		}
	]
}
`

	res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(firstRequestBody))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusCreated, res.StatusCode)

	duplicateRequestBody := `
{
	"team_name": "backend", 
	"members": [
		{
			"user_id": "u2",
			"username": "Bob",
			"is_active": true
		}
	]
}
`

	res, err = s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(duplicateRequestBody))
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
    "message": "backend already exists"
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

	res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(invalidJSON))
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
			res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(tc.requestBody))
			s.Require().NoError(err)
			defer res.Body.Close()

			s.Require().Equal(tc.expectedCode, res.StatusCode)
		})
	}
}

func (s *TeamAddSuite) TestAddTeamUpdatesExistingUsers() {
	firstRequestBody := `
{
	"team_name": "team1",
	"members": [
		{
			"user_id": "u1",
			"username": "OldName",
			"is_active": true
		},
		{
			"user_id": "u2",
			"username": "User2",
			"is_active": false
		}
	]
}
`

	res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(firstRequestBody))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusCreated, res.StatusCode)

	secondRequestBody := `
{
	"team_name": "team1",
	"members": [
		{
			"user_id": "u1",
			"username": "NewName",
			"is_active": false
		},
		{
			"user_id": "u2",
			"username": "NewName2",
			"is_active": true
		}
	]
}
`

	res, err = s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(secondRequestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Equal("team1", response.TeamName)
	s.Require().Len(response.Members, 2)

	s.Require().Equal("u1", response.Members[0].ID)
	s.Require().Equal("NewName", response.Members[0].Name)
	s.Require().False(response.Members[0].IsActive)

	s.Require().Equal("u2", response.Members[1].ID)
	s.Require().Equal("NewName2", response.Members[1].Name)
	s.Require().True(response.Members[1].IsActive)
}

func (s *TeamAddSuite) TestAddTeamUpdatesOnlyOneExistingUsers() {
	firstRequestBody := `
{
	"team_name": "team1",
	"members": [
		{
			"user_id": "u1",
			"username": "OldName",
			"is_active": true
		},
		{
			"user_id": "u2",
			"username": "OldUser2",
			"is_active": false
		}
	]
}
`

	res, err := s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(firstRequestBody))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusCreated, res.StatusCode)

	secondRequestBody := `
{
	"team_name": "team1",
	"members": [
		{
			"user_id": "u1",
			"username": "NewName",
			"is_active": false
		}
	]
}
`

	res, err = s.server.Client().Post(s.server.URL+"/team/add", "", bytes.NewBufferString(secondRequestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := teamsHandler.AddTeamResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Equal("team1", response.TeamName)
	s.Require().Len(response.Members, 2)

	s.Require().Equal("u1", response.Members[0].ID)
	s.Require().Equal("NewName", response.Members[0].Name)
	s.Require().False(response.Members[0].IsActive)

	s.Require().Equal("u2", response.Members[1].ID)
	s.Require().Equal("OldUser2", response.Members[1].Name)
	s.Require().False(response.Members[1].IsActive)
}
