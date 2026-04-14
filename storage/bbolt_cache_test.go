package storage

import (
	"path/filepath"
	"testing"

	bbolt "go.etcd.io/bbolt"
	bbolterrors "go.etcd.io/bbolt/errors"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// flushAll implements testableCacheManager for BboltCacheManager.
func (cache *BboltCacheManager) flushAll() error {
	return cache.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range []string{NSStatus, NSLatest} {
			if err := tx.DeleteBucket([]byte(name)); err != nil && err != bbolterrors.ErrBucketNotFound {
				return err
			}
			if _, err := tx.CreateBucket([]byte(name)); err != nil {
				return err
			}
		}
		return nil
	})
}

// ─── status suite ────────────────────────────────────────────────────────────

type bboltStatusSuite struct{ cacheStatusSuite }

func (s *bboltStatusSuite) SetupTest() {
	require.NoError(s.T(), s.mgr.flushAll())
	seedStatuses(s.T(), s.mgr)
}

func (s *bboltStatusSuite) TearDownSuite() {
	s.mgr.Close() //nolint:errcheck
}

func TestBboltCacheStatuses(t *testing.T) {
	mgr := &BboltCacheManager{path: filepath.Join(t.TempDir(), "cache.db")}
	require.NoError(t, mgr.Start())
	suite.Run(t, &bboltStatusSuite{cacheStatusSuite{mgr: mgr}})
}

// ─── latest suite ────────────────────────────────────────────────────────────

type bboltLatestSuite struct{ cacheLatestSuite }

func (s *bboltLatestSuite) TearDownSuite() {
	s.mgr.Close() //nolint:errcheck
}

func TestBboltCacheLatest(t *testing.T) {
	mgr := &BboltCacheManager{path: filepath.Join(t.TempDir(), "cache.db")}
	require.NoError(t, mgr.Start())
	suite.Run(t, &bboltLatestSuite{cacheLatestSuite{mgr: mgr}})
}
