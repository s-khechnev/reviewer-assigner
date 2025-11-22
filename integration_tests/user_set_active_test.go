package integration_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	"reviewer-assigner/internal/http/handlers/users"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

type UserSetActiveSuite struct {
	BaseSuite
}

func (s *UserSetActiveSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *UserSetActiveSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *UserSetActiveSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/user_set_active"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestUserSetActiveSuite_Run(t *testing.T) {
	suite.Run(t, new(UserSetActiveSuite))
}

func (s *UserSetActiveSuite) TestDefault() {
	requestBody := `
{
  "user_id": "u1_Alice",
  "is_active": false
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/users/setIsActive", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	var response users.SetIsActiveResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
  "user": {
    "user_id": "u1_Alice",
    "username": "Alice",
    "team_name": "payments",
    "is_active": false
  }
}
`

	JSONEq(s.T(), expected, response)
}

func (s *UserSetActiveSuite) TestUserNotFound() {
	requestBody := `
{
  "user_id": "user_not_found_id",
  "is_active": false
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/users/setIsActive", "", bytes.NewBufferString(requestBody))
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

func (s *UserSetActiveSuite) TestValidation() {
	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
	}{
		{
			name:         "empty",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{}
`,
		},
		{
			name:         "missing_is_active",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
  "user_id": "u2"
}
`,
		},
		{
			name:         "missing_user_id",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
  "is_active": true
}
`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.server.
				Client().
				Post(s.server.URL+"/users/setIsActive", "", bytes.NewBufferString(tc.requestBody))
			s.Require().NoError(err)
			defer res.Body.Close()

			s.Require().Equal(tc.expectedCode, res.StatusCode)
		})
	}
}
