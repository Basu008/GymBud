package app

import (
	"fmt"

	"github.com/Basu008/GymBud/app/exercise"
	"github.com/Basu008/GymBud/app/media"
	"github.com/Basu008/GymBud/app/routine"
	"github.com/Basu008/GymBud/app/social"
	"github.com/Basu008/GymBud/app/user"
	"github.com/Basu008/GymBud/app/workout"
	mongorepo "github.com/Basu008/GymBud/db/mongo"
	"github.com/Basu008/GymBud/db/postgres"
)

func InitService(a *App) error {
	postgresDB := a.Postgres

	userRepo, err := postgres.NewUserRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init user repository: %w", err)
	}
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
	socialRepo, err := postgres.NewSocialRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init social repository: %w", err)
	}
	mediaRepo, err := postgres.NewMediaRepo(postgresDB)
	if err != nil {
		return fmt.Errorf("init media repository: %w", err)
	}
	a.UserService = user.NewUserService(&user.Opts{
		Repo:        userRepo,
		SocialRepo:  socialRepo,
		MediaRepo:   mediaRepo,
		Logger:      a.Logger,
		AuthService: a.AuthService,
	})
	a.ExerciseService = exercise.NewExerciseService(&exercise.Opts{
		Repo:   exerciseRepo,
		Config: a.Config,
		Logger: a.Logger,
	})
	a.RoutineService = routine.NewRoutineService(&routine.Opts{
		Repo:       routineRepo,
		SocialRepo: socialRepo,
		Logger:     a.Logger,
	})
	a.SocialService = social.NewSocialService(&social.Opts{
		Repo:     socialRepo,
		UserRepo: userRepo,
		Logger:   a.Logger,
	})
	a.WorkoutService = workout.NewWorkoutService(&workout.Opts{
		Repo:        workoutRepo,
		RoutineRepo: routineRepo,
		SocialRepo:  socialRepo,
		UserRepo:    userRepo,
		Logger:      a.Logger,
	})
	a.MediaService = media.NewMediaService(&media.Opts{
		Firebase: a.Firebase,
		Repo:     mediaRepo,
		Logger:   a.Logger,
	})

	return nil
}
