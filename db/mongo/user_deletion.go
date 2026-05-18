package mongo

import (
	"context"
	"fmt"
	"time"

	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
)

type UserDeletionRepo struct {
	collection *mongodriver.Collection
}

type userDeletionDocument struct {
	UserID    string    `bson:"user_id"`
	DeletedAt time.Time `bson:"deleted_at"`
	Reason    string    `bson:"reason"`
}

func NewUserDeletionRepo(db *mongodriver.Database) (*UserDeletionRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return &UserDeletionRepo{
		collection: db.Collection("user_deletions"),
	}, nil
}

func (r *UserDeletionRepo) LogDeletion(ctx context.Context, userID, reason string) error {
	doc := userDeletionDocument{
		UserID:    userID,
		DeletedAt: time.Now().UTC(),
		Reason:    reason,
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("log user deletion: %w", err)
	}
	return nil
}
