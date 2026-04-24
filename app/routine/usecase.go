package routine

import (
	modelroutine "github.com/Basu008/GymBud/model/routine"
	modelsocial "github.com/Basu008/GymBud/model/social"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo       modelroutine.Repository
	SocialRepo modelsocial.Repository
	Logger     *zerolog.Logger
}

type Service struct {
	repo       modelroutine.Repository
	socialRepo modelsocial.Repository
	logger     *zerolog.Logger
}

func NewRoutineService(opts *Opts) *Service {
	return &Service{
		repo:       opts.Repo,
		socialRepo: opts.SocialRepo,
		logger:     opts.Logger,
	}
}
