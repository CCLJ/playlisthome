package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the shared connection pool used across the app.
var Pool *pgxpool.Pool

// Connect initialises the Postgres connection pool.
// It reads DATABASE_URL from the environment and retries a few times
// so the app starts cleanly even if Postgres is still warming up.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing DATABASE_URL: %w", err)
	}

	// Sensible pool defaults — tune for your workload
	cfg.MaxConns = 25
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	// Retry loop — Docker Compose healthcheck should handle ordering,
	// but an extra retry here is cheap insurance.
	const maxAttempts = 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			log.Printf("[db] connect attempt %d/%d failed: %v", attempt, maxAttempts, err)
		} else if err = pool.Ping(ctx); err != nil {
			pool.Close()
			log.Printf("[db] ping attempt %d/%d failed: %v", attempt, maxAttempts, err)
		} else {
			log.Printf("[db] connected to Postgres (pool max=%d)", cfg.MaxConns)
			Pool = pool
			return pool, nil
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}

	return nil, fmt.Errorf("could not connect to Postgres after %d attempts", maxAttempts)
}
