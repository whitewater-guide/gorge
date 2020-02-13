package storage

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSqlite(t *testing.T) {
	sqliteMgr, err := NewSqliteDb(0)
	if err != nil {
		t.Fatalf("failed to init sqlite manager: %v", err)
	}
	sqliteSuite := &DbTestSuite{mgr: &(sqliteMgr.DbManager)}
	suite.Run(t, sqliteSuite)
}
