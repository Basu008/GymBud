package exercise

import "time"

const ExercisesTable = "exercises"

type Exercise struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Category      string    `db:"category" json:"category"`
	IsMadeByAdmin bool      `db:"is_made_by_admin" json:"isMadeByAdmin"`
	Equipment     string    `db:"equipment" json:"equipment"`
	MovementMode  *string   `db:"movement_mode" json:"movement_mode,omitempty"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}
