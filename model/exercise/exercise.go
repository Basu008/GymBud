package exercise

import "time"

const ExercisesTable = "exercises"

type Exercise struct {
	ID               string    `db:"id" json:"id"`
	Name             string    `db:"name" json:"name"`
	Slug             string    `db:"slug" json:"slug"`
	Category         string    `db:"category" json:"category"`
	UserID           *string   `db:"user_id" json:"user_id,omitempty"`
	IsMadeByAdmin    bool      `db:"is_made_by_admin" json:"-"`
	Equipment        string    `db:"equipment" json:"equipment"`
	PrimaryMuscle    string    `db:"primary_muscle" json:"primary_muscle"`
	SecondaryMuscles []string  `db:"secondary_muscles" json:"secondary_muscles"`
	Difficulty       string    `db:"difficulty" json:"difficulty"`
	MovementMode     *string   `db:"movement_mode" json:"movement_mode,omitempty"`
	IsActive         bool      `db:"is_active" json:"is_active"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}
