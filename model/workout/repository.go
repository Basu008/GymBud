package workout

import "context"

type Repository interface {
	Create(ctx context.Context, workout *Workout) error
	GetByID(ctx context.Context, workoutID string) (*Workout, error)
	LikeWorkout(ctx context.Context, workoutID, userID string) (*Workout, error)
	UnlikeWorkout(ctx context.Context, workoutID, userID string) (*Workout, error)
	GetLatestPersonalRecord(ctx context.Context, userID, exerciseID string) (*PersonalRecord, error)
	CreatePersonalRecord(ctx context.Context, record *PersonalRecord) error
}
