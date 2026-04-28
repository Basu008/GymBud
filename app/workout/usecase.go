package workout

import (
	modelroutine "github.com/Basu008/GymBud/model/routine"
	modelworkout "github.com/Basu008/GymBud/model/workout"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo        modelworkout.Repository
	RoutineRepo modelroutine.Repository
	Logger      *zerolog.Logger
}

type Service struct {
	repo        modelworkout.Repository
	routineRepo modelroutine.Repository
	logger      *zerolog.Logger
}

func NewWorkoutService(opts *Opts) *Service {
	return &Service{
		repo:        opts.Repo,
		routineRepo: opts.RoutineRepo,
		logger:      opts.Logger,
	}
}
