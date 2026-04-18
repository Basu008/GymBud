package app

import (
	"github.com/Basu008/GymBud/server/config"
	"github.com/rs/zerolog"
)

type App struct {
	Logger *zerolog.Logger
	Config *config.Config
}

type Options struct {
	Logger *zerolog.Logger
	Config *config.Config
}

func NewApp(opts *Options) *App {
	return &App{
		Logger: opts.Logger,
		Config: opts.Config,
	}
}
