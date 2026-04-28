package schema

import "time"

type CreateWorkoutSetInput struct {
	SetNumber      int     `json:"set_number" validate:"required"`
	ActualReps     int     `json:"actual_reps" validate:"required"`
	ActualWeightKG float64 `json:"actual_weight_kg"`
}

type CreateWorkoutExerciseInput struct {
	ExerciseID string                  `json:"exercise_id" validate:"required"`
	Sets       []CreateWorkoutSetInput `json:"sets" validate:"required"`
}

type CreateWorkoutBody struct {
	RoutineID  string                       `json:"routine_id" validate:"required"`
	StartTime  time.Time                    `json:"start_time" validate:"required"`
	EndTime    time.Time                    `json:"end_time" validate:"required"`
	Visibility string                       `json:"visibility" validate:"required"`
	Notes      *string                      `json:"notes"`
	Exercises  []CreateWorkoutExerciseInput `json:"exercises" validate:"required"`
}

type WorkoutPRFlagsPayload struct {
	WeightPR       bool `json:"weight_pr"`
	RepPR          bool `json:"rep_pr"`
	Estimated1RMPR bool `json:"estimated_1rm_pr"`
}

type WorkoutSetPayload struct {
	SetNumber       int                   `json:"set_number"`
	PlannedMinReps  int                   `json:"planned_min_reps"`
	PlannedMaxReps  int                   `json:"planned_max_reps"`
	PlannedWeightKG *float64              `json:"planned_weight_kg,omitempty"`
	ActualReps      int                   `json:"actual_reps"`
	ActualWeightKG  float64               `json:"actual_weight_kg"`
	PRFlags         WorkoutPRFlagsPayload `json:"pr_flags"`
}

type WorkoutExercisePayload struct {
	ExerciseID   string               `json:"exercise_id"`
	ExerciseName string               `json:"exercise_name"`
	OrderIndex   int                  `json:"order_index"`
	Sets         []*WorkoutSetPayload `json:"sets"`
}

type WorkoutStatsPayload struct {
	TotalSets   int     `json:"total_sets"`
	TotalReps   int     `json:"total_reps"`
	TotalVolume float64 `json:"total_volume"`
	PRCount     int     `json:"pr_count"`
}

type WorkoutPayload struct {
	ID          string                    `json:"id"`
	UserID      string                    `json:"user_id"`
	RoutineID   string                    `json:"routine_id"`
	Title       string                    `json:"title"`
	StartedAt   time.Time                 `json:"started_at"`
	EndedAt     time.Time                 `json:"ended_at"`
	DurationSec int                       `json:"duration_sec"`
	Visibility  string                    `json:"visibility"`
	Notes       *string                   `json:"notes,omitempty"`
	HasPR       bool                      `json:"has_pr"`
	LikesCount  int                       `json:"likes_count"`
	LikedByMe   bool                      `json:"liked_by_me"`
	Exercises   []*WorkoutExercisePayload `json:"exercises"`
	Stats       WorkoutStatsPayload       `json:"stats"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

type WorkoutResponse struct {
	Workout *WorkoutPayload `json:"workout"`
}

type DeleteWorkoutResponse struct {
	DeletedID string `json:"deleted_id"`
}

type PaginationPayload struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type WorkoutsResponse struct {
	Workouts   []*WorkoutPayload   `json:"workouts"`
	Pagination PaginationPayload `json:"pagination"`
}
