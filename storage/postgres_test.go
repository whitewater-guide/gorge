//go:build nodocker
// +build nodocker

package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

var pgConnStr string

func TestMain(m *testing.M) {
	var db *sql.DB
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not create dockertest pool: %s", err)
	}
	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("timescale/timescaledb", "1.5.1-pg11", []string{"POSTGRES_PASSWORD=postgres"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		pgConnStr = fmt.Sprintf("postgres://postgres:postgres@localhost:%s/postgres?sslmode=disable", resource.GetPort("5432/tcp"))
		db, err = sql.Open("postgres", pgConnStr)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	db.Close()
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestPostgres(t *testing.T) {
	pgMgr, err := NewPostgresManager(pgConnStr, 0)
	if err != nil {
		log.Fatalf("failed to init sqlite manager: %v", err)
	}
	pgSuite := &DbTestSuite{mgr: &(pgMgr.DbManager)}
	suite.Run(t, pgSuite)
}
