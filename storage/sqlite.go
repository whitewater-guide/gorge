package storage

import (
	"embed"
	"fmt"
	"math"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/sqlite/*.sql
var sqliteFS embed.FS

// SqliteManager implements DatabaseManager interface for Sqlite datbase https://github.com/mattn/go-sqlite3
type SqliteManager struct {
	DbManager
}

// NewSqliteDb creates SqliteManager with given chunkSize
// SqliteManager cannot be used for write access concurrently. Test usage only
func NewSqliteDb(chunkSize int) *SqliteManager {
	return &SqliteManager{
		DbManager{
			defaultStart:     "datetime('now', '-30 days')",
			nearestDayClause: "ABS(julianday(timestamp) - julianday(%s))",
			saveChunkSize:    chunkSize,
		},
	}

}

// Start implements DatabaseManager interface
func (mgr *SqliteManager) Start() error {
	db, err := obtainConnection("sqlite3", "file::memory:?cache=shared", 2, 60)
	if err != nil {
		return fmt.Errorf("failed to obtain sqlite connection: %w", err)
	}
	// See SQLite FAQ: https://github.com/mattn/go-sqlite3#faq
	db.SetConnMaxLifetime(time.Duration(math.MaxInt64))
	mgr.db = db

	// Run migrations
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration db driver: %w", err)
	}

	d, err := iofs.New(sqliteFS, "migrations/sqlite")
	if err != nil {
		return fmt.Errorf("failed to create migration iofs source: %w", err)
	}

	migrations, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
