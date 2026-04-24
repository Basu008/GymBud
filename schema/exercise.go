package schema

import "time"

type CreateExerciseBody struct {
	Name         string  `json:"name" validate:"required"`
	Category     string  `json:"category" validate:"required"`
	Equipment    string  `json:"equipment" validate:"required"`
	MovementMode *string `json:"movement_mode"`
	IsAdmin      *bool   `json:"is_admin"`
}

type UpdateExerciseBody struct {
	Name         *string `json:"name"`
	Category     *string `json:"category"`
	Equipment    *string `json:"equipment"`
	MovementMode *string `json:"movement_mode"`
	IsAdmin      *bool   `json:"is_admin"`
	IsActive     *bool   `json:"is_active"`
}

type ExercisePayload struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Equipment    string    `json:"equipment"`
	MovementMode *string   `json:"movement_mode,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ExerciseResponse struct {
	Exercise *ExercisePayload `json:"exercise"`
}

type ExercisesResponse struct {
	Exercises []*ExercisePayload `json:"exercises"`
}

type DeleteExerciseResponse struct {
	DeletedID string `json:"deleted_id"`
}
