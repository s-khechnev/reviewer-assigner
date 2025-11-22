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

type PullRequestMergeSuite struct {
	BaseSuite
}

func (s *PullRequestMergeSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *PullRequestMergeSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *PullRequestMergeSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/pull_request_merge"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestPullRequestMergeSuite_Run(t *testing.T) {
	suite.Run(t, new(PullRequestMergeSuite))
}

func (s *PullRequestMergeSuite) TestMergeDefault() {
	requestBody := `
{
  "pull_request_id": "pr_opened_id"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/merge", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := prHandler.MergePullRequestResponse{}
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
    "pull_request_id": "pr_opened_id",
    "pull_request_name": "Opened PR",
    "author_id": "u1_Alice",
    "status": "MERGED",
    "assigned_reviewers": [
      "{{.id0}}",
      "{{.id1}}"
    ],
	"created_at": "{{.createdAt}}",
	"merged_at": "{{.mergedAt}}"
  }
}
`

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
		"id0":       response.AssignedReviewers[0],
		"id1":       response.AssignedReviewers[1],
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
		"mergedAt":  template.HTML(response.MergedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}

func (s *PullRequestMergeSuite) TestNotFound() {
	requestBody := `
{
  "pull_request_id": "pr_not_found_id"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/merge", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusNotFound, res.StatusCode)

	response := handlers.ErrorResponse{}
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

func (s *PullRequestMergeSuite) TestAlreadyMerged() {
	requestBody := `
{
  "pull_request_id": "pr_merged_id"
}
`

	res, err := s.server.Client().
		Post(s.server.URL+"/pullRequest/merge", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := prHandler.MergePullRequestResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Len(response.AssignedReviewers, 2)

	expectedReviewers := map[string]struct{}{
		"u3_John":  {},
		"u1_Alice": {},
	}
	for _, reviewer := range response.AssignedReviewers {
		_, ok := expectedReviewers[reviewer]
		s.Require().True(ok)
	}

	expectedTemplate := `
{
  "pr": {
    "pull_request_id": "pr_merged_id",
    "pull_request_name": "Merged PR",
    "author_id": "u2_Bob",
    "status": "MERGED",
    "assigned_reviewers": [
      "{{.id0}}",
      "{{.id1}}"
    ],
	"created_at": "{{.createdAt}}",
	"merged_at": "{{.mergedAt}}"
  }
}
`

	alreadyMergedAt, err := time.Parse(time.DateTime, "2024-01-15 10:33:00")
	s.Require().NoError(err)

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]any{
		"id0":       response.AssignedReviewers[0],
		"id1":       response.AssignedReviewers[1],
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
		"mergedAt":  template.HTML(alreadyMergedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}
