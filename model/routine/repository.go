package routine

import (
	"context"
	"errors"
)

var ErrRoutineNotFound = errors.New("routine not found")
var ErrRoutineExerciseNotFound = errors.New("one or more exercises not found")

type ListFilter struct {
	UserID string
	Offset int64
	Limit  int64
}

type Repository interface {
	CountByUserID(ctx context.Context, userID string) (int, error)
	Create(ctx context.Context, routine *Routine) (*Routine, error)
	ListByUserID(ctx context.Context, filter *ListFilter) ([]*Routine, int64, error)
	GetByID(ctx context.Context, userID, routineID string) (*Routine, error)
	GetByRoutineID(ctx context.Context, routineID string) (*Routine, error)
	ReplaceByID(ctx context.Context, userID string, routine *Routine) (*Routine, error)
	DeleteByID(ctx context.Context, userID, routineID string) error
}
