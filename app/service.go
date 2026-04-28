package app

import (
	"fmt"

	"github.com/Basu008/GymBud/app/exercise"
	"github.com/Basu008/GymBud/app/routine"
	"github.com/Basu008/GymBud/app/social"
	"github.com/Basu008/GymBud/app/user"
	"github.com/Basu008/GymBud/app/workout"
	mongorepo "github.com/Basu008/GymBud/db/mongo"
	"github.com/Basu008/GymBud/db/postgres"
)

func InitService(a *App) error {
	postgresDB := a.Postgres

	exerciseRepo, err := postgres.NewExerciseRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init exercise repository: %w", err)
	}
	routineRepo, err := postgres.NewRoutineRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init routine repository: %w", err)
	}
	workoutRepo, err := mongorepo.NewWorkoutRepo(a.Mongo)
	if err != nil {
		return fmt.Errorf("init workout repository: %w", err)
	}
	userRepo, err := postgres.NewUserRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init user repository: %w", err)
	}
	socialRepo, err := postgres.NewSocialRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init social repository: %w", err)
	}
	a.ExerciseService = exercise.NewExerciseService(&exercise.Opts{
		Repo:   exerciseRepo,
		Logger: a.Logger,
	})
	a.RoutineService = routine.NewRoutineService(&routine.Opts{
		Repo:       routineRepo,
		SocialRepo: socialRepo,
		Logger:     a.Logger,
	})
	a.UserService = user.NewUserService(&user.Opts{
		Repo:        userRepo,
		SocialRepo:  socialRepo,
		Logger:      a.Logger,
		AuthService: a.AuthService,
	})
	a.SocialService = social.NewSocialService(&social.Opts{
		Repo:     socialRepo,
		UserRepo: userRepo,
		Logger:   a.Logger,
	})
	a.WorkoutService = workout.NewWorkoutService(&workout.Opts{
		Repo:        workoutRepo,
		RoutineRepo: routineRepo,
		Logger:      a.Logger,
	})

	return nil
}
