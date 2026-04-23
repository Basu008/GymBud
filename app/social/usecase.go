package social

import (
	modelsocial "github.com/Basu008/GymBud/model/social"
	modeluser "github.com/Basu008/GymBud/model/user"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo     modelsocial.Repository
	UserRepo modeluser.Repository
	Logger   *zerolog.Logger
}

type Service struct {
	repo     modelsocial.Repository
	userRepo modeluser.Repository
	logger   *zerolog.Logger
}

func NewSocialService(opts *Opts) *Service {
	return &Service{
		repo:     opts.Repo,
		userRepo: opts.UserRepo,
		logger:   opts.Logger,
	}
}
