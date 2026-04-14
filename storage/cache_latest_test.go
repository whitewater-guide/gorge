package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type clTestSuite struct{ cacheLatestSuite }

func (s *clTestSuite) TearDownSuite() {
	s.mgr.Close() //nolint:errcheck
}

func TestCacheLatest(t *testing.T) {
	mgr := &EmbeddedCacheManager{}
	require.NoError(t, mgr.Start())
	suite.Run(t, &clTestSuite{cacheLatestSuite{mgr: mgr}})
}
