package storage

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSqlite(t *testing.T) {
	mgr := NewSqliteDb(0)
	mgr.Start() //nolint:errcheck
	tests := &DbTestSuite{mgr: &(mgr.DbManager)}
	suite.Run(t, tests)
}
