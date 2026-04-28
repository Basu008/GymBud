package api

import (
	"errors"
	"net/http"

	appworkout "github.com/Basu008/GymBud/app/workout"
	"github.com/Basu008/GymBud/schema"
	"github.com/Basu008/GymBud/server/handler"
)

func (a *API) createWorkout(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.CreateWorkoutBody
	if err := handler.BindJSONStrict(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &body) {
		return
	}

	response, err := a.App.WorkoutService.CreateWorkout(r.Context(), ctx.UserClaim.UserID, &body)
	if err != nil {
		switch {
		case errors.Is(err, appworkout.ErrRoutineNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, appworkout.ErrRoutineExerciseNotFound):
			handler.BadRequest(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.Created(w, response)
}

func (a *API) likeWorkout(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	workoutID := pathID(r)
	response, err := a.App.WorkoutService.LikeWorkout(r.Context(), ctx.UserClaim.UserID, workoutID)
	if err != nil {
		if errors.Is(err, appworkout.ErrWorkoutNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) unlikeWorkout(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	workoutID := pathID(r)
	response, err := a.App.WorkoutService.UnlikeWorkout(r.Context(), ctx.UserClaim.UserID, workoutID)
	if err != nil {
		if errors.Is(err, appworkout.ErrWorkoutNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}
