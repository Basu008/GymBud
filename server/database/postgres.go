package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Basu008/GymBud/server/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(cfg config.PostgresDatabaseConfig) *pgxpool.Pool {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("failed to parse postgres config: %s", err.Error())
	}

	if cfg.MaxConnections > 0 {
		poolConfig.MaxConns = cfg.MaxConnections
	}
	if cfg.MinConnections > 0 {
		poolConfig.MinConns = cfg.MinConnections
	}
	if cfg.HealthCheckTime > 0 {
		poolConfig.HealthCheckPeriod = cfg.HealthCheckTime
	}

	timeout := cfg.ConnectTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("failed to create postgres pool: %s", err.Error())
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatalf("failed to connect to postgres: %s", err.Error())
	}

	fmt.Println("connected to postgres")

	return pool
}
