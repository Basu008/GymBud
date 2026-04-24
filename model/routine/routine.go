package routine

import (
	"time"

	modelexercise "github.com/Basu008/GymBud/model/exercise"
)

type Routine struct {
	ID        string
	UserID    string
	Name      string
	Exercises []*RoutineExercise
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RoutineExercise struct {
	ID         string
	RoutineID  string
	ExerciseID string
	OrderIndex int
	Notes      *string
	Exercise   *modelexercise.Exercise
	Sets       []*RoutineExerciseSet
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type RoutineExerciseSet struct {
	ID                string
	RoutineExerciseID string
	SetNumber         int
	MinReps           int
	MaxReps           int
	TargetWeightKG    *float64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
