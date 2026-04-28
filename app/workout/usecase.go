package workout

import (
	modelsocial "github.com/Basu008/GymBud/model/social"
	modelroutine "github.com/Basu008/GymBud/model/routine"
	modeluser "github.com/Basu008/GymBud/model/user"
	modelworkout "github.com/Basu008/GymBud/model/workout"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo        modelworkout.Repository
	RoutineRepo modelroutine.Repository
	SocialRepo  modelsocial.Repository
	UserRepo    modeluser.Repository
	Logger      *zerolog.Logger
}

type Service struct {
	repo        modelworkout.Repository
	routineRepo modelroutine.Repository
	socialRepo  modelsocial.Repository
	userRepo    modeluser.Repository
	logger      *zerolog.Logger
}

func NewWorkoutService(opts *Opts) *Service {
	return &Service{
		repo:        opts.Repo,
		routineRepo: opts.RoutineRepo,
		socialRepo:  opts.SocialRepo,
		userRepo:    opts.UserRepo,
		logger:      opts.Logger,
	}
}
