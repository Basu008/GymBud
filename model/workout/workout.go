package workout

import (
	"errors"
	"time"
)

var ErrWorkoutNotFound = errors.New("workout not found")

type Workout struct {
	ID          string             `bson:"_id" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	RoutineID   string             `bson:"routine_id" json:"routine_id"`
	Title       string             `bson:"title" json:"title"`
	StartedAt   time.Time          `bson:"started_at" json:"started_at"`
	EndedAt     time.Time          `bson:"ended_at" json:"ended_at"`
	DurationSec int                `bson:"duration_sec" json:"duration_sec"`
	Visibility  string             `bson:"visibility" json:"visibility"`
	Notes       *string            `bson:"notes,omitempty" json:"notes,omitempty"`
	Exercises   []*WorkoutExercise `bson:"exercises" json:"exercises"`
	Stats       WorkoutStats       `bson:"stats" json:"stats"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type ListFilter struct {
	UserID       string
	Visibility   *string
	StartedAtGTE *time.Time
	StartedAtLT  *time.Time
	Offset       int64
	Limit        int64
}

type WorkoutExercise struct {
	ExerciseID   string                `bson:"exercise_id" json:"exercise_id"`
	ExerciseName string                `bson:"exercise_name" json:"exercise_name"`
	OrderIndex   int                   `bson:"order_index" json:"order_index"`
	Sets         []*WorkoutExerciseSet `bson:"sets" json:"sets"`
}

type WorkoutExerciseSet struct {
	SetNumber       int      `bson:"set_number" json:"set_number"`
	PlannedMinReps  int      `bson:"planned_min_reps" json:"planned_min_reps"`
	PlannedMaxReps  int      `bson:"planned_max_reps" json:"planned_max_reps"`
	PlannedWeightKG *float64 `bson:"planned_weight_kg,omitempty" json:"planned_weight_kg,omitempty"`
	ActualReps      int      `bson:"actual_reps" json:"actual_reps"`
	ActualWeightKG  float64  `bson:"actual_weight_kg" json:"actual_weight_kg"`
	PRFlags         PRFlags  `bson:"pr_flags" json:"pr_flags"`
}

type PRFlags struct {
	WeightPR       bool `bson:"weight_pr" json:"weight_pr"`
	RepPR          bool `bson:"rep_pr" json:"rep_pr"`
	Estimated1RMPR bool `bson:"estimated_1rm_pr" json:"estimated_1rm_pr"`
}

type WorkoutStats struct {
	TotalSets   int     `bson:"total_sets" json:"total_sets"`
	TotalReps   int     `bson:"total_reps" json:"total_reps"`
	TotalVolume float64 `bson:"total_volume" json:"total_volume"`
	PRCount     int     `bson:"pr_count" json:"pr_count"`
}

type PersonalRecord struct {
	ID           string    `bson:"_id"`
	UserID       string    `bson:"user_id"`
	ExerciseID   string    `bson:"exercise_id"`
	ExerciseName string    `bson:"exercise_name"`
	BestWeightKG float64   `bson:"best_weight_kg"`
	BestReps     int       `bson:"best_reps"`
	Estimated1RM float64   `bson:"estimated_1rm"`
	WorkoutID    string    `bson:"workout_id"`
	UpdatedAt    time.Time `bson:"updated_at"`
}
