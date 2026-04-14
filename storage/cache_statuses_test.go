package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	obo = "1d141432-5b1a-4a37-ab8e-78d912631fe5" // one-by-one job id

	aOk      = "622fd526-7fea-4870-be00-60c06576c5d5" // all-at-once job id (success)
	aErr     = "e2fb933f-cf9b-4956-a389-b95f7e2bba32" // all-at-once job id (success in past, then error)
	aErrOnly = "58cd6558-f356-450f-8ea6-755894059cd0" // all-at-once job id (error only)
)

var now = time.Date(2019, time.January, 1, 12, 0, 0, 0, time.UTC)

type csTestSuite struct{ cacheStatusSuite }

func (s *csTestSuite) SetupTest() {
	require.NoError(s.T(), s.mgr.flushAll())
	seedStatuses(s.T(), s.mgr)
}

func (s *csTestSuite) TearDownSuite() {
	s.mgr.Close() //nolint:errcheck
}

func TestCacheStatuses(t *testing.T) {
	mgr := &EmbeddedCacheManager{}
	require.NoError(t, mgr.Start())
	suite.Run(t, &csTestSuite{cacheStatusSuite{mgr: mgr}})
}
