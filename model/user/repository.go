package user

import (
	"context"
	"errors"
)

var ErrUsernameAlreadyExists = errors.New("username already exists")

type Repository interface {
	Create(ctx context.Context, user *User) error
}
