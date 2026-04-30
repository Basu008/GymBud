package database

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"github.com/Basu008/GymBud/server/config"
	"google.golang.org/api/option"
)

func NewFireBaseApp(fc *config.FirebaseConfig) *firebase.App {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{
		StorageBucket: fc.StorageBucket,
	}, option.WithAuthCredentialsJSON(option.ServiceAccount, fc.ReturnConfigJSON()))
	if err != nil {
		log.Fatalf("failed to init firebase: %v", err)
	}
	return app
}
