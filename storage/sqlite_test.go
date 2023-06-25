package storage

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

func TestSqlite(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mgr := NewSqliteDb(logrus.NewEntry(logger), 0)
	mgr.Start() //nolint:errcheck
	tests := &DbTestSuite{mgr: &(mgr.DbManager)}
	suite.Run(t, tests)
}
