package integration_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
	"html/template"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	prHandler "reviewer-assigner/internal/http/handlers/pull_requests"
	"testing"
	"time"
)

type PullRequestCreateSuite struct {
	BaseSuite
}

func (s *PullRequestCreateSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *PullRequestCreateSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *PullRequestCreateSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/pull_request_create"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestPullRequestCreateSuite_Run(t *testing.T) {
	suite.Run(t, new(PullRequestCreateSuite))
}

func (s *PullRequestCreateSuite) TestCreateDefault() {
	requestBody := `
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search",
  "author_id": "u1_Alice"
}
`

	res, err := s.server.Client().Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := prHandler.CreatePullRequestResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Len(response.AssignedReviewers, 2)

	expectedReviewers := map[string]struct{}{
		"u3_John": {},
		"u2_Bob":  {},
	}
	for _, reviewer := range response.AssignedReviewers {
		_, ok := expectedReviewers[reviewer]
		s.Require().True(ok)
	}

	expectedTemplate := `
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1_Alice",
    "status": "OPEN",
    "assigned_reviewers": [
      "{{.id0}}",
      "{{.id1}}"
    ],
	"created_at": "{{.createdAt}}"
  }
}
`

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]interface{}{
		"id0":       response.AssignedReviewers[0],
		"id1":       response.AssignedReviewers[1],
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestCreateSuite) TestCreateNotFoundAuthor() {
	requestBody := `
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search",
  "author_id": "u1_NOT_FOUND"
}
`

	res, err := s.server.Client().Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusNotFound, res.StatusCode)

	var errorResp handlers.ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&errorResp)
	s.Require().NoError(err)

	expected := `
{
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found"
  }
}
`

	JSONEq(s.T(), expected, errorResp)
}

func (s *PullRequestCreateSuite) TestCreateAlreadyExists() {
	requestBody := `
{
  "pull_request_id": "pr_already_exists_id",
  "pull_request_name": "Implement payment gateway integration",
  "author_id": "u1_Alice"
}
`

	res, err := s.server.Client().Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusConflict, res.StatusCode)

	var errorResp handlers.ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&errorResp)
	s.Require().NoError(err)

	expected := `
{
  "error": {
    "code": "PR_EXISTS",
    "message": "PR pr_already_exists_id already exists"
  }
}
`

	JSONEq(s.T(), expected, errorResp)
}

func (s *TeamAddSuite) TestCreateMissingRequiredFields() {
	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
	}{
		{
			name:         "missing_author_id",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search"
}
`,
		},
		{
			name:         "missing_pull_request_name",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
  "pull_request_id": "pr-1001",
  "author_id": "u1"
}
`,
		},
		{
			name:         "missing_pull_request_id",
			expectedCode: http.StatusUnprocessableEntity,
			requestBody: `
{
  "pull_request_name": "Add search",
  "author_id": "u1"
}
`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.server.Client().Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(tc.requestBody))
			s.Require().NoError(err)
			defer res.Body.Close()

			s.Require().Equal(tc.expectedCode, res.StatusCode)
		})
	}
}
