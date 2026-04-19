package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Basu008/GymBud/server/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthUser struct {
	UserID    string
	Role      string
	Plan      string
	SessionID string
}

func (u *AuthUser) IsAdmin() bool {
	return u != nil && u.Role == "admin"
}

func (u *AuthUser) IsPremium() bool {
	return u != nil && u.Plan == "premium"
}

func (u *AuthUser) CanAccessPremium() bool {
	return u != nil && (u.IsPremium() || u.IsAdmin())
}

type Session struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
	Plan         string    `json:"plan"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AuthService struct {
	jwtManager *JWTManager
	redis      *redis.Client
	// postgres        *pgxpool.Pool
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Options struct {
	Config   *config.Config
	Redis    *redis.Client
	Postgres *pgxpool.Pool
}

func NewAuthService(opts *Options) *AuthService {
	tokenAuthConfig := opts.Config.TokenAuthConfig
	return &AuthService{
		jwtManager: NewJWTManager(tokenAuthConfig.JWTSecret, tokenAuthConfig.JWTIssuer),
		redis:      opts.Redis,
		// postgres:        opts.Postgres,
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: 7 * 24 * time.Hour,
	}
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *AuthService) AuthenticateRequest(token string) (*AuthUser, error) {
	claims, err := s.jwtManager.ParseToken(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type")
	}

	val, err := s.redis.Get(context.Background(), sessionKey(claims.SessionID)).Result()
	if err != nil {
		return nil, errors.New("session not found")
	}

	var session Session
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}

	if session.UserID != claims.UserID || session.Role != claims.Role || session.Plan != claims.Plan {
		return nil, errors.New("session mismatch")
	}

	return &AuthUser{
		UserID:    claims.UserID,
		Role:      claims.Role,
		Plan:      claims.Plan,
		SessionID: claims.SessionID,
	}, nil
}
