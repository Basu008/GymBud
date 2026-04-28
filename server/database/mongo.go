package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Basu008/GymBud/server/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoClient(cfg config.MongoDatabaseConfig) *mongo.Client {
	if cfg.URI == "" {
		log.Fatal("failed to connect to mongo: uri is required")
	}

	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		log.Fatalf("failed to create mongo client: %s", err.Error())
	}

	if err := client.Ping(ctx, nil); err != nil {
		if closeErr := client.Disconnect(ctx); closeErr != nil {
			log.Printf("failed to disconnect mongo client after ping error: %v", closeErr)
		}
		log.Fatalf("failed to connect to mongo: %s", err.Error())
	}

	fmt.Println("connected to mongo")

	return client
}
