package app

import (
	"github.com/Basu008/GymBud/app/user"
	"github.com/Basu008/GymBud/db/postgres"
)

func InitService(a *App) {
	postgresDB := a.Postgres

	a.UserService = user.NewUserService(&user.Opts{
		Repo:   postgres.NewUserRepo(postgresDB),
		Logger: a.Logger,
	})
}
