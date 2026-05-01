package user

import (
	"context"
	"errors"
)

var ErrUsernameAlreadyExists = errors.New("username already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrBodyMetricsNotFound = errors.New("body metrics not found")
var ErrUserInactive = errors.New("user is inactive")

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdateByID(ctx context.Context, userID string, updates *UserUpdate) (*User, error)
	UpdatePrivacyByID(ctx context.Context, userID string, isPrivate bool) (*User, error)
	UpdateActiveByID(ctx context.Context, userID string, isActive bool) (*User, error)
	CreateBodyMetrics(ctx context.Context, metrics *BodyMetrics) (*CurrentStats, error)
	GetCurrentBodyMetrics(ctx context.Context, userID string) (*BodyMetrics, error)
	ListBodyMetrics(ctx context.Context, userID string, limit int) ([]*BodyMetrics, error)
	DeleteBodyMetrics(ctx context.Context, userID, metricsID string) (*CurrentStats, error)
}
