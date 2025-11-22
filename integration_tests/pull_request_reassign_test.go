package integration_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
	"html/template"
	"net/http"
	prHandler "reviewer-assigner/internal/http/handlers/pull_requests"
	"testing"
	"time"
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

func (s *PullRequestReassignSuite) TestReassignDefault() {
	requestBody := `
{
  "pull_request_id": "pr_opened_id",
  "old_reviewer_id": "u2_Bob"
}
`

	res, err := s.server.Client().Post(s.server.URL+"/pullRequest/reassign", "", bytes.NewBufferString(requestBody))
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

	expected := s.loader.LoadTemplate(expectedTemplate, map[string]interface{}{
		"id0":       response.AssignedReviewers[0],
		"id1":       response.AssignedReviewers[1],
		"createdAt": template.HTML(response.CreatedAt.Format(time.RFC3339Nano)),
	})

	JSONEq(s.T(), expected, response)
}
