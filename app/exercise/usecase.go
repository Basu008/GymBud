package exercise

import (
	modelexercise "github.com/Basu008/GymBud/model/exercise"
	"github.com/Basu008/GymBud/server/config"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo   modelexercise.Repository
	Config *config.Config
	Logger *zerolog.Logger
}

type Service struct {
	repo   modelexercise.Repository
	config *config.Config
	logger *zerolog.Logger
}

func NewExerciseService(opts *Opts) *Service {
	return &Service{
		repo:   opts.Repo,
		config: opts.Config,
		logger: opts.Logger,
	}
}
