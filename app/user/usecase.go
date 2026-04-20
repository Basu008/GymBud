package user

import (
	"github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/server/auth"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo        user.Repository
	Logger      *zerolog.Logger
	AuthService *auth.AuthService
}

type Service struct {
	repo        user.Repository
	logger      *zerolog.Logger
	authService *auth.AuthService
}

func NewUserService(opts *Opts) *Service {
	return &Service{
		repo:        opts.Repo,
		logger:      opts.Logger,
		authService: opts.AuthService,
	}
}
