package app

import (
	"github.com/Basu008/GymBud/app/user"
	"github.com/Basu008/GymBud/server/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type App struct {
	Logger   *zerolog.Logger
	Config   *config.Config
	Postgres *pgxpool.Pool
	Redis    *redis.Client

	//Services
	UserService *user.Service
}

type Options struct {
	Logger   *zerolog.Logger
	Config   *config.Config
	Postgres *pgxpool.Pool
	Redis    *redis.Client
}

func NewApp(opts *Options) *App {
	return &App{
		Logger:   opts.Logger,
		Config:   opts.Config,
		Postgres: opts.Postgres,
		Redis:    opts.Redis,
	}
}
