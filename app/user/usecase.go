package user

import (
	"github.com/Basu008/GymBud/model/user"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo   user.Repository
	Logger *zerolog.Logger
}

type Service struct {
	repo   user.Repository
	logger *zerolog.Logger
}

func NewUserService(opts *Opts) *Service {
	return &Service{
		repo:   opts.Repo,
		logger: opts.Logger,
	}
}
