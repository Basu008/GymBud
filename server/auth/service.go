package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	Plan         string    `json:"plan"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuthService struct {
	jwtManager *JWTManager
	redis      *redis.Client
	// postgres        *pgxpool.Pool
}

var ErrSessionExpired = errors.New("session expired, please log in again")
var ErrLoginRequired = errors.New("login required")

type LoginSession struct {
	SessionID    string `json:"session_id"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
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

	accessToken, err := s.jwtManager.GenerateToken(userID, plan, sessionID, "access")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := Session{
		SessionID:    sessionID,
		UserID:       userID,
		Plan:         plan,
		RefreshToken: refreshToken,
		CreatedAt:    now,
	}

	rawSession, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Set(ctx, sessionKey(sessionID), rawSession, 0).Err(); err != nil {
		return nil, err
	}

	return &LoginSession{
		SessionID:    sessionID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

func (s *AuthService) AuthenticateRequest(token string) (*AuthUser, error) {
	claims, err := s.jwtManager.ParseToken(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, ErrInvalidAccessToken
	}

	val, err := s.redis.Get(context.Background(), sessionKey(claims.SessionID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionExpired
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}

	if session.UserID != claims.UserID || session.Plan != claims.Plan {
		return nil, ErrInvalidAccessToken
	}

	return &AuthUser{
		UserID:    claims.UserID,
		Plan:      claims.Plan,
		SessionID: claims.SessionID,
	}, nil
}

func (s *AuthService) LogoutSession(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return errors.New("session id is required")
	}

	return s.redis.Del(ctx, sessionKey(sessionID)).Err()
}
