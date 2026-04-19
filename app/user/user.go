package user

import (
	"context"
	"errors"
	"strings"

	"github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/schema"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) SignUpUser(ctx context.Context, opts *schema.SignUpUserBody) error {
	if opts == nil {
		return errors.New("signup payload is required")
	}

	username := strings.TrimSpace(opts.Username)
	email := strings.TrimSpace(opts.Email)
	password := strings.TrimSpace(opts.Password)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := user.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		DisplayName:  opts.DisplayName,
	}

	return s.repo.Create(ctx, &user)
}
