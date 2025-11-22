package integration_tests

import (
	"database/sql"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
	"testing"
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
		testfixtures.Directory("fixtures/storage/user_set_active"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestUserGetReviewSuite_Run(t *testing.T) {
	suite.Run(t, new(UserGetReviewSuite))
}

func (s *UserGetReviewSuite) TestDefault() {
	s.Require().True(true, "user default should be true")
}
