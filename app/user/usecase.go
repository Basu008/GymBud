package user

import (
	modelsocial "github.com/Basu008/GymBud/model/social"
	"github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/server/auth"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo        user.Repository
	SocialRepo  modelsocial.Repository
	Logger      *zerolog.Logger
	AuthService *auth.AuthService
}

type Service struct {
	repo        user.Repository
	socialRepo  modelsocial.Repository
	logger      *zerolog.Logger
	authService *auth.AuthService
}

func NewUserService(opts *Opts) *Service {
	return &Service{
		repo:        opts.Repo,
		socialRepo:  opts.SocialRepo,
		logger:      opts.Logger,
		authService: opts.AuthService,
	}
}
