package storage

import (
	"embed"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/config"

	// postgres
	_ "github.com/lib/pq"
)

//go:embed migrations/postgres/*.sql
var pgFS embed.FS

// PostgresManager implements DatabaseManager interface
type PostgresManager struct {
	DbManager
	logger    *logrus.Entry
	pgConnStr string
	// pgConnStr without password for logging purposes
	censoredConnStr string
}

// NewPostgresManager creates new PostgresManager with connection string and chunk size
func newPostgresManager(logger *logrus.Entry, cfg *config.Config) *PostgresManager {
	return &PostgresManager{
		DbManager: DbManager{
			defaultStart:     "NOW() - interval '30 days'",
			nearestDayClause: "abs(extract(epoch from timestamp - %s::timestamptz))",
			saveChunkSize:    cfg.DbChunkSize,
		},
		pgConnStr: fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			cfg.Pg.User,
			url.QueryEscape(cfg.Pg.Password),
			cfg.Pg.Host,
			cfg.Pg.Db,
		),
		censoredConnStr: fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			cfg.Pg.User,
			"**********",
			cfg.Pg.Host,
			cfg.Pg.Db,
		),
		logger: logger,
	}
}

// Start implements DatabaseManager interface
func (mgr *PostgresManager) Start() error {
	mgr.logger.Debugf("connecting to %s", mgr.censoredConnStr)
	pg, err := obtainConnection("postgres", mgr.pgConnStr, 2, 60)
	if err != nil {
		return fmt.Errorf("failed to init postgres: %w", err)
	}
	mgr.db = pg

	// Run migrations
	driver, err := postgres.WithInstance(pg.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration db driver: %w", err)
	}

	d, err := iofs.New(pgFS, "migrations/postgres")
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
