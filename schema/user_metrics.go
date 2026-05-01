package schema

import (
	modeluser "github.com/Basu008/GymBud/model/user"
)

type CreateBodyMetricsBody struct {
	HeightCM *float64 `json:"height_cm"`
	WeightKG *float64 `json:"weight_kg"`
	Source   string   `json:"source"`
}

type BodyMetricsResponse struct {
	Metrics      *modeluser.BodyMetrics  `json:"metrics"`
	CurrentStats *modeluser.CurrentStats `json:"current_stats"`
}

type BodyMetricsListResponse struct {
	Metrics    []*modeluser.BodyMetrics `json:"metrics"`
	Pagination PaginationPayload        `json:"pagination"`
}

type DeleteBodyMetricsResponse struct {
	DeletedID    string                  `json:"deleted_id"`
	CurrentStats *modeluser.CurrentStats `json:"current_stats"`
}
