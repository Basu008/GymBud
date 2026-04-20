package user

import (
	"time"

	"github.com/google/uuid"
)

type BodyMetrics struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	HeightCM   *float64  `db:"height_cm" json:"height_cm,omitempty"`
	WeightKG   *float64  `db:"weight_kg" json:"weight_kg,omitempty"`
	RecordedAt time.Time `db:"recorded_at" json:"recorded_at"`
	Source     string    `db:"source" json:"source"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type CurrentStats struct {
	UserID          uuid.UUID `db:"user_id" json:"user_id"`
	CurrentHeightCM *float64  `db:"current_height_cm" json:"current_height_cm,omitempty"`
	CurrentWeightKG *float64  `db:"current_weight_kg" json:"current_weight_kg,omitempty"`
	BMI             *float64  `db:"bmi" json:"bmi,omitempty"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
