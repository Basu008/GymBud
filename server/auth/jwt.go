package auth

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID    string `json:"uid"`
	Plan      string `json:"plan"`
	SessionID string `json:"sid"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	issuer string
}

var ErrAccessTokenExpired = errors.New("access token expired")
var ErrInvalidAccessToken = errors.New("invalid access token")

func NewJWTManager(secret, issuer string) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
	}
}

func (j *JWTManager) GenerateToken(userID, plan, sessionID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := UserClaims{
		UserID:    userID,
		Plan:      plan,
		SessionID: sessionID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.secret)
	if err != nil {
		return "", err
	}

	return encodeToken(signedToken), nil
}

func (j *JWTManager) ParseToken(tokenStr string) (*UserClaims, error) {
	decodedToken, err := decodeToken(tokenStr)
	if err != nil {
		return nil, ErrInvalidAccessToken
	}

	token, err := jwt.ParseWithClaims(string(decodedToken), &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrAccessTokenExpired
		}
		return nil, ErrInvalidAccessToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidAccessToken
	}

	return claims, nil
}

func encodeToken(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func decodeToken(token string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(token)
}
