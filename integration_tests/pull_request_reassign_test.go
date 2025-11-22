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

type PullRequestReassignSuite struct {
	BaseSuite
}

func (s *PullRequestReassignSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *PullRequestReassignSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *PullRequestReassignSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/pull_request_reassign"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestPullRequestReassignSuite_Run(t *testing.T) {
	suite.Run(t, new(PullRequestReassignSuite))
}

func (s *PullRequestReassignSuite) TestDefault() {
	requestBody := `
{
  "pull_request_id": "pr_opened_id",
  "old_reviewer_id": "u2_Bob"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/reassign", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := prHandler.ReassignPullRequestResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Len(response.AssignedReviewers, 2)

	expectedReviewers := map[string]struct{}{
		"u3_John": {},
		"u4_Mike": {},
	}
	for _, reviewer := range response.AssignedReviewers {
		_, ok := expectedReviewers[reviewer]
		s.Require().True(ok)
	}

	expectedTemplate := `
{
  "pr": {
    "pull_request_id": "pr_opened_id",
    "pull_request_name": "Opened PR",
    "author_id": "u1_Alice",
    "status": "OPEN",
    "assigned_reviewers": [
      "{{.id0}}",
      "{{.id1}}"
    ],
	"created_at": "{{.createdAt}}"
  },
  "replaced_by": "u4_Mike"
}
`

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
		"id0":       response.AssignedReviewers[0],
		"id1":       response.AssignedReviewers[1],
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestReassignSuite) TestNotFound() {
	notFoundResp := `
{
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found"
  }
}
`

	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "PR not found",
			requestBody: `
{
  "pull_request_id": "pr_not_found_id",
  "old_reviewer_id": "u2_Bob"
}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  notFoundResp,
		},
		{
			name: "Old reviewer not found",
			requestBody: `
{
  "pull_request_id": "pr_opened_id",
  "old_reviewer_id": "u_not_found_id"
}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  notFoundResp,
		},
		{
			name: "PR and old reviewer not found",
			requestBody: `
{
  "pull_request_id": "pr_not_found_id",
  "old_reviewer_id": "u_not_found_id"
}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  notFoundResp,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.server.
				Client().
				Post(s.server.URL+"/pullRequest/reassign", "", bytes.NewBufferString(tc.requestBody))
			s.Require().NoError(err)
			defer res.Body.Close()

			s.Require().Equal(tc.expectedStatus, res.StatusCode)

			response := handlers.ErrorResponse{}
			err = json.NewDecoder(res.Body).Decode(&response)
			s.Require().NoError(err)

			JSONEq(s.T(), tc.expectedError, response)
		})
	}
}

func (s *PullRequestReassignSuite) TestMerged() {
	requestBody := `
{
  "pull_request_id": "pr_merged_id",
  "old_reviewer_id": "u2_Bob"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/reassign", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusConflict, res.StatusCode)

	var response handlers.ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
  "error": {
    "code": "PR_MERGED",
    "message": "cannot reassign on merged PR"
  }
}
`

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestReassignSuite) TestNotAssigned() {
	requestBody := `
{
  "pull_request_id": "pr_opened_id",
  "old_reviewer_id": "u4_Mike"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/reassign", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusConflict, res.StatusCode)

	var response handlers.ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
  "error": {
    "code": "NOT_ASSIGNED",
    "message": "reviewer is not assigned to this PR"
  }
}
`

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestReassignSuite) TestNoCandidates() {
	requestBody := `
{
  "pull_request_id": "pr_no_candidates_for_reassign",
  "old_reviewer_id": "infra_Azat"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/reassign", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusConflict, res.StatusCode)

	var response handlers.ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := `
{
  "error": {
    "code": "NO_CANDIDATE",
    "message": "no active replacement candidate in team"
  }
}
`

	JSONEq(s.T(), expected, response)
}
