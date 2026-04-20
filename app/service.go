package app

import (
	"fmt"

	"github.com/Basu008/GymBud/app/user"
	"github.com/Basu008/GymBud/db/postgres"
)

func InitService(a *App) error {
	postgresDB := a.Postgres

	userRepo, err := postgres.NewUserRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init user repository: %w", err)
	}
	a.UserService = user.NewUserService(&user.Opts{
		Repo:        userRepo,
		Logger:      a.Logger,
		AuthService: a.AuthService,
	})

	return nil
}
