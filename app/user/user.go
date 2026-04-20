package user

import (
	"context"
	"errors"
	"strings"

	modeluser "github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/schema"
	"golang.org/x/crypto/bcrypt"
)

var ErrUsernameAlreadyExists = errors.New("username already exists")
var ErrInvalidCredentials = errors.New("invalid username or password")

func (s *Service) SignUpUser(ctx context.Context, opts *schema.SignUpUserBody) error {
	username := strings.TrimSpace(opts.Username)
	email := strings.TrimSpace(opts.Email)
	password := strings.TrimSpace(opts.Password)
	displayName := strings.TrimSpace(opts.DisplayName)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := modeluser.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		DisplayName:  displayName,
		Plan:         "free",
		IsActive:     true,
		IsVerified:   false,
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		if errors.Is(err, modeluser.ErrUsernameAlreadyExists) {
			return ErrUsernameAlreadyExists
		}
		return err
	}

	return nil
}

func (s *Service) LoginUser(ctx context.Context, opts *schema.LoginUserBody) (*schema.LoginUserResponse, error) {
	username := strings.TrimSpace(opts.Username)
	password := strings.TrimSpace(opts.Password)

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	session, err := s.authService.CreateLoginSession(ctx, user.ID.String(), user.Plan)
	if err != nil {
		return nil, err
	}

	return &schema.LoginUserResponse{
		User:            user,
		RefreshToken:    session.RefreshToken,
		AccessToken:     session.AccessToken,
		AccessTokenTTL:  session.AccessTokenTTL,
		RefreshTokenTTL: session.RefreshTokenTTL,
	}, nil
}
