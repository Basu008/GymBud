package schema

import "time"

type RoutineSetInput struct {
	SetNumber      int      `json:"set_number"`
	MinReps        int      `json:"min_reps"`
	MaxReps        int      `json:"max_reps"`
	TargetWeightKG *float64 `json:"target_weight_kg,omitempty"`
}

type RoutineExerciseInput struct {
	ExerciseID string            `json:"exercise_id"`
	OrderIndex int               `json:"order_index"`
	Notes      *string           `json:"notes"`
	Sets       []RoutineSetInput `json:"sets"`
}

type CreateRoutineBody struct {
	Name      string                 `json:"name" validate:"required"`
	Exercises []RoutineExerciseInput `json:"exercises"`
}

type UpdateRoutineBody struct {
	Name      *string                 `json:"name"`
	Exercises *[]RoutineExerciseInput `json:"exercises"`
}

type RoutineSetPayload struct {
	ID                string    `json:"id"`
	RoutineExerciseID string    `json:"routine_exercise_id"`
	SetNumber         int       `json:"set_number"`
	MinReps           int       `json:"min_reps"`
	MaxReps           int       `json:"max_reps"`
	TargetWeightKG    *float64  `json:"target_weight_kg,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RoutineExercisePayload struct {
	ID         string               `json:"id"`
	RoutineID  string               `json:"routine_id"`
	ExerciseID string               `json:"exercise_id"`
	OrderIndex int                  `json:"order_index"`
	Notes      *string              `json:"notes,omitempty"`
	Exercise   *ExercisePayload     `json:"exercise,omitempty"`
	Sets       []*RoutineSetPayload `json:"sets"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

type RoutinePayload struct {
	ID        string                    `json:"id"`
	UserID    string                    `json:"user_id"`
	Name      string                    `json:"name"`
	Exercises []*RoutineExercisePayload `json:"exercises"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

type RoutineResponse struct {
	Routine *RoutinePayload `json:"routine"`
}

type RoutinesResponse struct {
	Routines []*RoutinePayload `json:"routines"`
}

type DeleteRoutineResponse struct {
	DeletedID string `json:"deleted_id"`
}
