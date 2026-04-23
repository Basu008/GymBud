package exercise

import (
	modelexercise "github.com/Basu008/GymBud/model/exercise"
	"github.com/rs/zerolog"
)

type Opts struct {
	Repo   modelexercise.Repository
	Logger *zerolog.Logger
}

type Service struct {
	repo   modelexercise.Repository
	logger *zerolog.Logger
}

func NewExerciseService(opts *Opts) *Service {
	return &Service{
		repo:   opts.Repo,
		logger: opts.Logger,
	}
}
