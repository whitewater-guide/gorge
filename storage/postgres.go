package storage

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// postgres
	_ "github.com/lib/pq"
)

//go:embed migrations/postgres/*.sql
var pgFS embed.FS

// PostgresManager implements DatabaseManager interface
type PostgresManager struct {
	DbManager
}

// NewPostgresManager creates new PostgresManager with connection string and chunk size
func NewPostgresManager(pgConnStr string, chunkSize int, withoutTimescale bool) (*PostgresManager, error) {
	manager := &PostgresManager{
		DbManager{
			defaultStart:     "NOW() - interval '30 days'",
			nearestDayClause: "abs(extract(epoch from timestamp - %s::timestamptz))",
			saveChunkSize:    chunkSize,
		},
	}

	pg, err := obtainConnection("postgres", pgConnStr, 2, 60)
	if err != nil {
		return nil, fmt.Errorf("failed to init postgres: %w", err)
	}
	manager.db = pg

	// Run migrations
	driver, err := postgres.WithInstance(pg.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration db driver: %w", err)
	}

	d, err := iofs.New(pgFS, "migrations/postgres")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration iofs source: %w", err)
	}

	migrations, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration: %w", err)
	}

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return manager, nil
}
