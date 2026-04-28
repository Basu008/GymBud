package app

import (
	"github.com/Basu008/GymBud/app/exercise"
	"github.com/Basu008/GymBud/app/routine"
	"github.com/Basu008/GymBud/app/social"
	"github.com/Basu008/GymBud/app/user"
	"github.com/Basu008/GymBud/server/auth"
	"github.com/Basu008/GymBud/server/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type App struct {
	Logger   *zerolog.Logger
	Config   *config.Config
	Mongo    *mongo.Database
	Postgres *pgxpool.Pool
	Redis    *redis.Client

	//Services
	ExerciseService *exercise.Service
	RoutineService  *routine.Service
	UserService     *user.Service
	SocialService   *social.Service
	AuthService     *auth.AuthService
}

type Options struct {
	Logger   *zerolog.Logger
	Config   *config.Config
	Mongo    *mongo.Database
	Postgres *pgxpool.Pool
	Redis    *redis.Client
}

func NewApp(opts *Options) *App {
	return &App{
		Logger:   opts.Logger,
		Config:   opts.Config,
		Mongo:    opts.Mongo,
		Postgres: opts.Postgres,
		Redis:    opts.Redis,
	}
}
