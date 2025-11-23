package integration_tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	statsHandler "reviewer-assigner/internal/http/handlers/stats"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

type StatsGetReviewersAssignmentsSuite struct {
	BaseSuite
}

func (s *StatsGetReviewersAssignmentsSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StatsGetReviewersAssignmentsSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *StatsGetReviewersAssignmentsSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/stats_reviewers_assignments"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestStatsGetReviewersAssignmentsSuite_Run(t *testing.T) {
	suite.Run(t, new(StatsGetReviewersAssignmentsSuite))
}

func (s *StatsGetReviewersAssignmentsSuite) TestDefault() {
	res, err := s.server.Client().Get(s.server.URL + "/stats/reviewers/assignments")
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := statsHandler.GetStatsUserAssignmentsResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := map[string]struct {
		Name  string
		Count int
	}{
		"u2_payments_reviewer": {
			Name:  "User2 Payments Reviewer",
			Count: 2,
		},
		"u5_backend_reviewer": {
			Name:  "User5 Backend Reviewer",
			Count: 2,
		},
		"u3_payments_reviewer_inactive": {
			Name:  "User3 Payments Reviewer Inactive",
			Count: 1,
		},
	}

	s.Require().Len(response.UserAssignments, len(expected))

	for _, assignment := range response.UserAssignments {
		expectAssignment := expected[assignment.UserID]

		s.Require().Equal(expectAssignment.Name, assignment.Name)
		s.Require().Equal(expectAssignment.Count, assignment.Count)
	}
}

func (s *StatsGetReviewersAssignmentsSuite) TestOnlyMerged() {
	res, err := s.server.Client().Get(s.server.URL + "/stats/reviewers/assignments?status=merged")
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := statsHandler.GetStatsUserAssignmentsResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := map[string]struct {
		Name  string
		Count int
	}{
		"u2_payments_reviewer": {
			Name:  "User2 Payments Reviewer",
			Count: 1,
		},
		"u5_backend_reviewer": {
			Name:  "User5 Backend Reviewer",
			Count: 1,
		},
	}

	s.Require().Len(response.UserAssignments, len(expected))

	for _, assignment := range response.UserAssignments {
		expectAssignment := expected[assignment.UserID]

		s.Require().Equal(expectAssignment.Name, assignment.Name)
		s.Require().Equal(expectAssignment.Count, assignment.Count)
	}
}

func (s *StatsGetReviewersAssignmentsSuite) TestOnlyActiveAndStatusCombination() {
	res, err := s.server.Client().
		Get(s.server.URL + "/stats/reviewers/assignments?status=open&active_only")
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := statsHandler.GetStatsUserAssignmentsResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expected := map[string]struct {
		Name  string
		Count int
	}{
		"u2_payments_reviewer": {
			Name:  "User2 Payments Reviewer",
			Count: 1,
		},
		"u5_backend_reviewer": {
			Name:  "User5 Backend Reviewer",
			Count: 1,
		},
	}

	s.Require().Len(response.UserAssignments, len(expected))

	for _, assignment := range response.UserAssignments {
		expectAssignment := expected[assignment.UserID]

		s.Require().Equal(expectAssignment.Name, assignment.Name)
		s.Require().Equal(expectAssignment.Count, assignment.Count)
	}
}
