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

func (s *Service) SignUpUser(ctx context.Context, opts *schema.SignUpUserBody) error {
	if opts == nil {
		return errors.New("signup payload is required")
	}

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
