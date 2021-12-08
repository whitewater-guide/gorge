package storage

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/config"
	"go.uber.org/fx"
)

func newDatabaseManager(lc fx.Lifecycle, cfg *config.Config, logger *logrus.Logger) (DatabaseManager, error) {
	var mgr DatabaseManager
	switch cfg.Db {
	case "postgres":
		mgr = newPostgresManager(cfg)
	case "inmemory":
		mgr = NewSqliteDb(cfg.DbChunkSize)
	default:
		return nil, fmt.Errorf("invalid database manager")
	}

	log := logger.WithField("logger", "db")

	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			log.Debug("starting")
			err := mgr.Start()
			log.Infof("started %s", cfg.Db)
			return err
		},
		OnStop: func(c context.Context) error {
			log.Debug("stopping")
			err := mgr.Close()
			log.Info("stopped")
			return err
		},
	})

	return mgr, nil
}

func newCacheManager(lc fx.Lifecycle, cfg *config.Config, logger *logrus.Logger) (CacheManager, error) {
	var mgr CacheManager
	switch cfg.Cache {
	case "redis":
		mgr = &RedisCacheManager{address: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)}
	case "inmemory":
		mgr = &EmbeddedCacheManager{}
	default:
		return nil, fmt.Errorf("invalid cache manager")
	}

	log := logger.WithField("logger", "cache")

	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			log.Debug("starting")
			err := mgr.Start()
			log.Infof("started %s", cfg.Cache)
			return err
		},
		OnStop: func(c context.Context) error {
			log.Debug("stopping")
			err := mgr.Close()
			log.Info("stopped")
			return err
		},
	})

	return mgr, nil
}

var Module = fx.Options(
	fx.Provide(newDatabaseManager),
	fx.Provide(newCacheManager),
)
