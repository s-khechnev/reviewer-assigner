package integration_tests

import (
	"database/sql"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
	"testing"
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
		testfixtures.Directory("fixtures/storage/team_get"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestUserSetActiveSuite_Run(t *testing.T) {
	suite.Run(t, new(UserSetActiveSuite))
}
