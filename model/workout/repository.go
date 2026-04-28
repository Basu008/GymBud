package workout

import "context"

type Repository interface {
	Create(ctx context.Context, workout *Workout) error
	GetLatestPersonalRecord(ctx context.Context, userID, exerciseID string) (*PersonalRecord, error)
	CreatePersonalRecord(ctx context.Context, record *PersonalRecord) error
}
