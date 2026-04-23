package exercise

import (
	"context"
	"errors"
)

var ErrExerciseNotFound = errors.New("exercise not found")
var ErrExerciseNameAlreadyExists = errors.New("exercise already exists for this equipment")
var ErrExerciseManagedByAdmin = errors.New("admin-created exercises cannot be updated or deleted")

type Repository interface {
	Create(ctx context.Context, exercise *Exercise) error
	GetByID(ctx context.Context, exerciseID string) (*Exercise, error)
	UpdateByID(ctx context.Context, exerciseID string, input *UpdateInput) (*Exercise, error)
	DeleteByID(ctx context.Context, exerciseID string) error
}
