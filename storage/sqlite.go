package storage

import (
	"fmt"
	"math"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"

	//
	_ "github.com/mattn/go-sqlite3"
	sqlite_migrations "github.com/whitewater-guide/gorge/storage/migrations/sqlite"
)

// SqliteManager implements DatabaseManager interface for Sqlite datbase https://github.com/mattn/go-sqlite3
type SqliteManager struct {
	DbManager
}

// NewSqliteDb creates SqliteManager with given chunkSize
// SqliteManager cannot be used for write access concurrently. Test usage only
func NewSqliteDb(chunkSize int) (*SqliteManager, error) {
	manager := &SqliteManager{
		DbManager{
			defaultStart:     "datetime('now', '-30 days')",
			nearestDayClause: "ABS(julianday(timestamp) - julianday(%s))",
			saveChunkSize:    chunkSize,
		},
	}

	db, err := obtainConnection("sqlite3", "file::memory:?cache=shared", 2, 60)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain sqlite connection: %w", err)
	}
	// See SQLite FAQ: https://github.com/mattn/go-sqlite3#faq
	db.SetConnMaxLifetime(time.Duration(math.MaxInt64))
	manager.db = db

	// Run migrations
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration db driver: %w", err)
	}

	src := bindata.Resource(sqlite_migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return sqlite_migrations.Asset(name)
		})

	d, err := bindata.WithInstance(src)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration bindata source: %w", err)
	}

	migrations, err := migrate.NewWithInstance("go-bindata", d, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration: %w", err)
	}

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return manager, nil
}
