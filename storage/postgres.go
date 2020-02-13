package storage

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"

	// postgres
	_ "github.com/lib/pq"
	postgres_migrations "github.com/whitewater-guide/gorge/storage/migrations/postgres"
)

// PostgresManager implements DatabaseManager interface
type PostgresManager struct {
	DbManager
}

// NewPostgresManager creates new PostgresManager with connection string and chunk size
func NewPostgresManager(pgConnStr string, chunkSize int) (*PostgresManager, error) {
	manager := &PostgresManager{
		DbManager{
			defaultStart:  "NOW() - interval '30 days'",
			saveChunkSize: chunkSize,
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

	src := bindata.Resource(postgres_migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return postgres_migrations.Asset(name)
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
