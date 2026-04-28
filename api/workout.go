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

func (a *API) getWorkoutByID(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	workoutID := pathID(r)
	response, err := a.App.WorkoutService.GetWorkoutByID(r.Context(), ctx.UserClaim.UserID, workoutID)
	if err != nil {
		switch {
		case errors.Is(err, appworkout.ErrWorkoutNotFound), errors.Is(err, appworkout.ErrUserNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, appworkout.ErrWorkoutAccessDenied):
			handler.Forbidden(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) getLatestWorkoutByRoutineID(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	routineID := pathID(r)
	response, err := a.App.WorkoutService.GetLatestWorkoutByRoutineID(r.Context(), ctx.UserClaim.UserID, routineID)
	if err != nil {
		switch {
		case errors.Is(err, appworkout.ErrRoutineNotFound), errors.Is(err, appworkout.ErrWorkoutNotFound):
			handler.NotFound(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) deleteWorkout(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	workoutID := pathID(r)
	response, err := a.App.WorkoutService.DeleteWorkout(r.Context(), ctx.UserClaim.UserID, workoutID)
	if err != nil {
		switch {
		case errors.Is(err, appworkout.ErrWorkoutNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, appworkout.ErrWorkoutDeleteNotAllowed):
			handler.Forbidden(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) listUserWorkouts(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	page, limit, err := a.paginationParams(r)
	if err != nil {
		handler.BadRequest(w, "page and limit must be positive integers")
		return
	}

	userID := pathID(r)
	response, err := a.App.WorkoutService.ListUserWorkouts(r.Context(), ctx.UserClaim.UserID, userID, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, appworkout.ErrUserNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, appworkout.ErrWorkoutAccessDenied):
			handler.Forbidden(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) listCurrentUserWorkouts(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	page, limit, err := a.paginationParams(r)
	if err != nil {
		handler.BadRequest(w, "page and limit must be positive integers")
		return
	}

	response, err := a.App.WorkoutService.ListUserWorkouts(r.Context(), ctx.UserClaim.UserID, ctx.UserClaim.UserID, page, limit)
	if err != nil {
		if errors.Is(err, appworkout.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
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
