package schema

import "time"

type CreateExerciseBody struct {
	Name             string   `json:"name" validate:"required"`
	Category         string   `json:"category" validate:"required"`
	Equipment        string   `json:"equipment" validate:"required"`
	PrimaryMuscle    string   `json:"primary_muscle" validate:"required"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Difficulty       string   `json:"difficulty" validate:"required"`
	MovementMode     *string  `json:"movement_mode"`
	IsAdmin          *bool    `json:"is_admin"`
}

type CreateExercisesBody struct {
	Exercises []CreateExerciseBody `json:"exercises" validate:"required"`
}

type UpdateExerciseBody struct {
	Name             *string   `json:"name"`
	Category         *string   `json:"category"`
	Equipment        *string   `json:"equipment"`
	PrimaryMuscle    *string   `json:"primary_muscle"`
	SecondaryMuscles *[]string `json:"secondary_muscles"`
	Difficulty       *string   `json:"difficulty"`
	MovementMode     *string   `json:"movement_mode"`
	IsAdmin          *bool     `json:"is_admin"`
	IsActive         *bool     `json:"is_active"`
}

type ExercisePayload struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	Category         string    `json:"category"`
	UserID           *string   `json:"user_id,omitempty"`
	Equipment        string    `json:"equipment"`
	PrimaryMuscle    string    `json:"primary_muscle"`
	SecondaryMuscles []string  `json:"secondary_muscles"`
	Difficulty       string    `json:"difficulty"`
	MovementMode     *string   `json:"movement_mode,omitempty"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ExerciseResponse struct {
	Exercise *ExercisePayload `json:"exercise"`
}

type ExercisesResponse struct {
	Exercises []*ExercisePayload `json:"exercises"`
}

type ExerciseCategoriesResponse struct {
	ExerciseCategories []string `json:"exerciseCategories"`
}

type ExerciseMusclesResponse struct {
	ExerciseMuscles []string `json:"exerciseMuscles"`
}

type ExerciseEquipmentsResponse struct {
	ExerciseEquipments []string `json:"exerciseEquipments"`
}

type ExerciseDifficultyResponse struct {
	ExerciseDifficulty []string `json:"exerciseDifficulty"`
}

type DeleteExerciseResponse struct {
	DeletedID string `json:"deleted_id"`
}
