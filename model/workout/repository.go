package workout

import "context"

type Repository interface {
	Create(ctx context.Context, workout *Workout) error
	GetByID(ctx context.Context, workoutID string) (*Workout, error)
	GetLatestByRoutineID(ctx context.Context, userID, routineID string) (*Workout, error)
	Delete(ctx context.Context, workoutID string) error
	ListByUserID(ctx context.Context, filter *ListFilter) ([]*Workout, int64, error)
	GetCurrentPRWorkoutIDs(ctx context.Context, userID string, workoutIDs []string) (map[string]bool, error)
	GetLatestPersonalRecord(ctx context.Context, userID, exerciseID string) (*PersonalRecord, error)
	CreatePersonalRecord(ctx context.Context, record *PersonalRecord) error
}
