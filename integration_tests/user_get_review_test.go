package integration_tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	usersHandler "reviewer-assigner/internal/http/handlers/users"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

type UserGetReviewSuite struct {
	BaseSuite
}

func (s *UserGetReviewSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *UserGetReviewSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *UserGetReviewSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage/user_get_review"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestUserGetReviewSuite_Run(t *testing.T) {
	suite.Run(t, new(UserGetReviewSuite))
}

func (s *UserGetReviewSuite) TestNotFoundUser() {
	const userID = "qweqweqweeq"

	res, err := s.server.Client().Get(s.server.URL + "/users/getReview?user_id=" + userID)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := usersHandler.GetReviewResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Empty(response.PullRequests)
}

func (s *UserGetReviewSuite) TestWithTwoOpenPRs() {
	const userID = "u5_Backend_Reviewer"

	res, err := s.server.Client().Get(s.server.URL + "/users/getReview?user_id=" + userID)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := usersHandler.GetReviewResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expectedPRIDs := map[string]struct{}{
		"pr_backend_1": {},
		"pr_backend_2": {},
	}

	s.Require().Len(response.PullRequests, len(expectedPRIDs))

	for _, pr := range response.PullRequests {
		_, ok := expectedPRIDs[pr.ID]
		s.Require().True(ok)
	}
}

func (s *UserGetReviewSuite) TestWithOneOpenOneMergedPRs() {
	const userID = "u6_Backend_Mixed"

	res, err := s.server.Client().Get(s.server.URL + "/users/getReview?user_id=" + userID)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := usersHandler.GetReviewResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expectedPRIDs := map[string]struct{}{
		"pr_backend_1": {},
		"pr_backend_3": {},
	}

	s.Require().Len(response.PullRequests, len(expectedPRIDs))

	for _, pr := range response.PullRequests {
		_, ok := expectedPRIDs[pr.ID]
		s.Require().True(ok)
	}
}

func (s *UserGetReviewSuite) TestInactiveUserOneMergedPR() {
	const userID = "u3_Payments_Inactive"

	res, err := s.server.Client().Get(s.server.URL + "/users/getReview?user_id=" + userID)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := usersHandler.GetReviewResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	expectedPRIDs := map[string]struct{}{
		"pr_payments_2": {},
	}

	s.Require().Len(response.PullRequests, len(expectedPRIDs))

	for _, pr := range response.PullRequests {
		_, ok := expectedPRIDs[pr.ID]
		s.Require().True(ok)
	}
}

func (s *UserGetReviewSuite) TestAuthorWithoutReview() {
	const userID = "u1_Payments_Author"

	res, err := s.server.Client().Get(s.server.URL + "/users/getReview?user_id=" + userID)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := usersHandler.GetReviewResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Empty(response.PullRequests)
}

func (s *UserGetReviewSuite) TestActiveUserWithoutReview() {
	const userID = "u7_Frontend_Only"

	res, err := s.server.Client().Get(s.server.URL + "/users/getReview?user_id=" + userID)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := usersHandler.GetReviewResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Require().Empty(response.PullRequests)
}
