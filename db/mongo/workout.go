package mongo

import (
	"context"
	"errors"
	"fmt"

	modelworkout "github.com/Basu008/GymBud/model/workout"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type WorkoutRepo struct {
	collection               *mongodriver.Collection
	personalRecordCollection *mongodriver.Collection
}

func NewWorkoutRepo(db *mongodriver.Database) (*WorkoutRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("mongo database is required")
	}

	return &WorkoutRepo{
		collection:               db.Collection("workouts"),
		personalRecordCollection: db.Collection("personal_records"),
	}, nil
}

func (r *WorkoutRepo) Create(ctx context.Context, workout *modelworkout.Workout) error {
	if _, err := r.collection.InsertOne(ctx, workout); err != nil {
		return fmt.Errorf("insert workout: %w", err)
	}

	return nil
}

func (r *WorkoutRepo) GetLatestPersonalRecord(ctx context.Context, userID, exerciseID string) (*modelworkout.PersonalRecord, error) {
	var record modelworkout.PersonalRecord
	err := r.personalRecordCollection.FindOne(
		ctx,
		bson.M{"user_id": userID, "exercise_id": exerciseID},
		options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: -1}}),
	).Decode(&record)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find personal record: %w", err)
	}

	return &record, nil
}

func (r *WorkoutRepo) CreatePersonalRecord(ctx context.Context, record *modelworkout.PersonalRecord) error {
	if _, err := r.personalRecordCollection.InsertOne(ctx, record); err != nil {
		return fmt.Errorf("insert personal record: %w", err)
	}

	return nil
}
