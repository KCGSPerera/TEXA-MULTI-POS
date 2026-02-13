package jobs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/cache"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/migration"
)

type Worker struct {
	db          *pgxpool.Pool
	cache       cache.Client
	databaseURL string
}

func NewWorker(db *pgxpool.Pool, cache cache.Client, databaseURL string) *Worker {
	return &Worker{db: db, cache: cache, databaseURL: databaseURL}
}

func (w *Worker) Start(ctx context.Context) {
	heartbeatTicker := time.NewTicker(60 * time.Second)
	consistencyTicker := time.NewTicker(6 * time.Hour)

	go func() {
		defer heartbeatTicker.Stop()
		defer consistencyTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.L().Info().Msg("background_worker_stopped")
				return
			case <-heartbeatTicker.C:
				w.logHeartbeat(ctx)
			case <-consistencyTicker.C:
				logger.L().Info().Msg("background_consistency_hook")
			}
		}
	}()
}

func (w *Worker) logHeartbeat(ctx context.Context) {
	dbOK := w.db.Ping(ctx) == nil
	redisOK := true
	if w.cache != nil && w.cache.Enabled() {
		redisOK = w.cache.Ping(ctx) == nil
	}

	version, err := migration.CurrentVersion(w.databaseURL)
	if err != nil {
		logger.L().Warn().Err(err).Msg("migration_version_check_failed")
	}

	logger.L().Info().
		Uint("migration_version", version).
		Bool("db_ok", dbOK).
		Bool("redis_ok", redisOK).
		Msg("system_heartbeat")
}
