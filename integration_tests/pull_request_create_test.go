package integration_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	prHandler "reviewer-assigner/internal/http/handlers/pullrequests"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
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

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
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

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
		"id0":       response.AssignedReviewers[0],
		"id1":       response.AssignedReviewers[1],
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestCreateSuite) TestCreateWithSingleReviewer() {
	requestBody := `
{
  "pull_request_id": "pr_with_single_reviewer_id",
  "pull_request_name": "PR with single reviewer",
  "author_id": "u4_Sarah"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := prHandler.CreatePullRequestResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Len(response.AssignedReviewers, 1)

	expectedReviewers := map[string]struct{}{
		"u5_Mike": {},
	}
	for _, reviewer := range response.AssignedReviewers {
		_, ok := expectedReviewers[reviewer]
		s.Require().True(ok)
	}

	expectedTemplate := `
{
  "pr": {
    "pull_request_id": "pr_with_single_reviewer_id",
    "pull_request_name": "PR with single reviewer",
    "author_id": "u4_Sarah",
    "status": "OPEN",
    "assigned_reviewers": [
      "{{.id0}}"
    ],
	"created_at": "{{.createdAt}}"
  }
}
`

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
		"id0":       response.AssignedReviewers[0],
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestCreateSuite) TestCreateWithNoReviewers() {
	requestBody := `
{
  "pull_request_id": "pr_with_no_reviewers_id",
  "pull_request_name": "PR with no reviewers",
  "author_id": "u6_Ivan"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := prHandler.CreatePullRequestResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Len(response.AssignedReviewers, 0)

	expectedTemplate := `
{
  "pr": {
    "pull_request_id": "pr_with_no_reviewers_id",
    "pull_request_name": "PR with no reviewers",
    "author_id": "u6_Ivan",
    "status": "OPEN",
    "assigned_reviewers": [],
	"created_at": "{{.createdAt}}"
  }
}
`

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestCreateSuite) TestCreateWithNoActiveMembers() {
	requestBody := `
{
  "pull_request_id": "pr_with_no_active_reviewers_id",
  "pull_request_name": "PR with no active reviewers",
  "author_id": "u7_ActiveAuthor"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := prHandler.CreatePullRequestResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Len(response.AssignedReviewers, 0)

	expectedTemplate := `
{
  "pr": {
    "pull_request_id": "pr_with_no_active_reviewers_id",
    "pull_request_name": "PR with no active reviewers",
    "author_id": "u7_ActiveAuthor",
    "status": "OPEN",
    "assigned_reviewers": [],
	"created_at": "{{.createdAt}}"
  }
}
`

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestCreateSuite) TestCreateWithOneActiveMembers() {
	requestBody := `
{
  "pull_request_id": "pr_with_one_active_reviewers_id",
  "pull_request_name": "PR with one active reviewers",
  "author_id": "u10_ActiveAuthor"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	response := prHandler.CreatePullRequestResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Len(response.AssignedReviewers, 1)

	expectedReviewers := map[string]struct{}{
		"u11_ActiveUser": {},
	}
	for _, reviewer := range response.AssignedReviewers {
		_, ok := expectedReviewers[reviewer]
		s.Require().True(ok)
	}

	expectedTemplate := `
{
  "pr": {
    "pull_request_id": "pr_with_one_active_reviewers_id",
    "pull_request_name": "PR with one active reviewers",
    "author_id": "u10_ActiveAuthor",
    "status": "OPEN",
    "assigned_reviewers": [
		"u11_ActiveUser"
	],
	"created_at": "{{.createdAt}}"
  }
}
`

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
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

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
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

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(requestBody))
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
			res, err := s.server.
				Client().
				Post(s.server.URL+"/pullRequest/create", "", bytes.NewBufferString(tc.requestBody))
			s.Require().NoError(err)
			defer res.Body.Close()

			s.Require().Equal(tc.expectedCode, res.StatusCode)
		})
	}
}
