package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/Basu008/GymBud/server/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.RedisConfig) *redis.Client {
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}

	opts := redis.Options{
		Addr:         cfg.Address,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.Database,
		DialTimeout:  connectTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	if cfg.TLS {
		opts.TLSConfig = &tls.Config{}
	}

	client := redis.NewClient(&opts)

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("failed to close redis client after ping error: %v", closeErr)
		}
		log.Fatalf("failed to connect to redis: %s", err.Error())
	}

	fmt.Println("connected to redis")

	return client
}
