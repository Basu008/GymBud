package app

import (
	"github.com/Basu008/GymBud/server/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type App struct {
	Logger   *zerolog.Logger
	Config   *config.Config
	Postgres *pgxpool.Pool
}

type Options struct {
	Logger   *zerolog.Logger
	Config   *config.Config
	Postgres *pgxpool.Pool
}

func NewApp(opts *Options) *App {
	return &App{
		Logger:   opts.Logger,
		Config:   opts.Config,
		Postgres: opts.Postgres,
	}
}
