package schema

import (
	modeluser "github.com/Basu008/GymBud/model/user"
)

type CreateBodyMetricsBody struct {
	HeightCM   *float64 `json:"height_cm"`
	WeightKG   *float64 `json:"weight_kg"`
	RecordedAt string   `json:"recorded_at" validate:"required"`
	Source     string   `json:"source"`
}

type BodyMetricsResponse struct {
	Metrics      *modeluser.BodyMetrics  `json:"metrics"`
	CurrentStats *modeluser.CurrentStats `json:"current_stats"`
}

type DeleteBodyMetricsResponse struct {
	DeletedID    string                  `json:"deleted_id"`
	CurrentStats *modeluser.CurrentStats `json:"current_stats"`
}
