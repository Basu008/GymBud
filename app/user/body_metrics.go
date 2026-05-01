package user

import (
	"context"
	"errors"
	"strings"
	"time"

	modeluser "github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/schema"
	"github.com/google/uuid"
)

var ErrBodyMetricsNotFound = errors.New("body metrics not found")

func (s *Service) CreateBodyMetrics(ctx context.Context, userID string, body *schema.CreateBodyMetricsBody) (*schema.BodyMetricsResponse, error) {
	parsedUserID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	source := strings.TrimSpace(body.Source)
	if source == "" {
		source = "manual"
	}

	metrics := &modeluser.BodyMetrics{
		ID:         uuid.New(),
		UserID:     parsedUserID,
		HeightCM:   body.HeightCM,
		WeightKG:   body.WeightKG,
		RecordedAt: time.Now().UTC(),
		Source:     source,
	}

	currentStats, err := s.repo.CreateBodyMetrics(ctx, metrics)
	if err != nil {
		return nil, err
	}

	return &schema.BodyMetricsResponse{
		Metrics:      metrics,
		CurrentStats: currentStats,
	}, nil
}

func (s *Service) GetCurrentBodyMetrics(ctx context.Context, userID string) (*schema.BodyMetricsResponse, error) {
	metrics, err := s.repo.GetCurrentBodyMetrics(ctx, strings.TrimSpace(userID))
	if err != nil {
		if errors.Is(err, modeluser.ErrBodyMetricsNotFound) {
			return nil, ErrBodyMetricsNotFound
		}
		return nil, err
	}

	return &schema.BodyMetricsResponse{
		Metrics: metrics,
	}, nil
}

func (s *Service) ListBodyMetrics(ctx context.Context, userID string, page, limit int) (*schema.BodyMetricsListResponse, error) {
	metrics, total, err := s.repo.ListBodyMetrics(ctx, strings.TrimSpace(userID), (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}

	return &schema.BodyMetricsListResponse{
		Metrics:    metrics,
		Pagination: schema.NewPaginationPayload(page, limit, total),
	}, nil
}

func (s *Service) DeleteBodyMetrics(ctx context.Context, userID, metricsID string) (*schema.DeleteBodyMetricsResponse, error) {
	parsedMetricsID, err := uuid.Parse(strings.TrimSpace(metricsID))
	if err != nil {
		return nil, err
	}

	currentStats, err := s.repo.DeleteBodyMetrics(ctx, strings.TrimSpace(userID), parsedMetricsID.String())
	if err != nil {
		if errors.Is(err, modeluser.ErrBodyMetricsNotFound) {
			return nil, ErrBodyMetricsNotFound
		}
		return nil, err
	}

	return &schema.DeleteBodyMetricsResponse{
		DeletedID:    parsedMetricsID.String(),
		CurrentStats: currentStats,
	}, nil
}
