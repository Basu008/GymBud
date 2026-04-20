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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthUser struct {
	UserID    string
	Plan      string
	SessionID string
}

func (u *AuthUser) IsPremium() bool {
	return u != nil && u.Plan == "premium"
}

func (u *AuthUser) CanAccessPremium() bool {
	return u != nil && u.IsPremium()
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

type LoginSession struct {
	SessionID        string    `json:"session_id"`
	RefreshToken     string    `json:"refresh_token"`
	AccessToken      string    `json:"access_token"`
	AccessTokenTTL   int64     `json:"access_token_ttl_seconds"`
	RefreshTokenTTL  int64     `json:"refresh_token_ttl_seconds"`
	SessionExpiresAt time.Time `json:"session_expires_at"`
}

type Options struct {
	Config   *config.Config
	Redis    *redis.Client
	Postgres *pgxpool.Pool
}

func NewAuthService(opts *Options) *AuthService {
	tokenAuthConfig := opts.Config.TokenAuthConfig
	accessTokenTTL := 15 * time.Minute
	if tokenAuthConfig.JWTExpiresAt != "" {
		if parsedTTL, err := time.ParseDuration(tokenAuthConfig.JWTExpiresAt); err == nil && parsedTTL > 0 {
			accessTokenTTL = parsedTTL
		}
	}
	return &AuthService{
		jwtManager: NewJWTManager(tokenAuthConfig.JWTSecret, tokenAuthConfig.JWTIssuer),
		redis:      opts.Redis,
		// postgres:        opts.Postgres,
		accessTokenTTL:  accessTokenTTL,
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

func (s *AuthService) CreateLoginSession(ctx context.Context, userID, plan string) (*LoginSession, error) {

	sessionID := uuid.NewString()
	refreshToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwtManager.GenerateToken(userID, plan, sessionID, "access", s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sessionExpiresAt := now.Add(s.refreshTokenTTL)
	session := Session{
		SessionID:    sessionID,
		UserID:       userID,
		Plan:         plan,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		ExpiresAt:    sessionExpiresAt,
	}

	rawSession, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Set(ctx, sessionKey(sessionID), rawSession, s.refreshTokenTTL).Err(); err != nil {
		return nil, err
	}

	return &LoginSession{
		SessionID:        sessionID,
		RefreshToken:     refreshToken,
		AccessToken:      accessToken,
		AccessTokenTTL:   int64(s.accessTokenTTL.Seconds()),
		RefreshTokenTTL:  int64(s.refreshTokenTTL.Seconds()),
		SessionExpiresAt: sessionExpiresAt,
	}, nil
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

	if session.UserID != claims.UserID || session.Plan != claims.Plan {
		return nil, errors.New("session mismatch")
	}

	return &AuthUser{
		UserID:    claims.UserID,
		Plan:      claims.Plan,
		SessionID: claims.SessionID,
	}, nil
}
