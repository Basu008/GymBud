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
	startedAtGTE, startedAtLT, err := a.workoutAnalyticsDateRange(r)
	if err != nil {
		handler.BadRequest(w, "start_date and end_date must be valid YYYY-MM-DD values, and end_date must be on or after start_date")
		return
	}

	userID := pathID(r)
	response, err := a.App.WorkoutService.ListUserWorkouts(r.Context(), ctx.UserClaim.UserID, userID, page, limit, startedAtGTE, startedAtLT)
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
	startedAtGTE, startedAtLT, err := a.workoutAnalyticsDateRange(r)
	if err != nil {
		handler.BadRequest(w, "start_date and end_date must be valid YYYY-MM-DD values, and end_date must be on or after start_date")
		return
	}

	response, err := a.App.WorkoutService.ListUserWorkouts(r.Context(), ctx.UserClaim.UserID, ctx.UserClaim.UserID, page, limit, startedAtGTE, startedAtLT)
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

func (a *API) listFollowingWorkouts(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	page, limit, err := a.paginationParams(r)
	if err != nil {
		handler.BadRequest(w, "page and limit must be positive integers")
		return
	}
	startedAtGTE, startedAtLT, err := a.workoutAnalyticsDateRange(r)
	if err != nil {
		handler.BadRequest(w, "start_date and end_date must be valid YYYY-MM-DD values, and end_date must be on or after start_date")
		return
	}

	response, err := a.App.WorkoutService.ListFollowingWorkouts(r.Context(), ctx.UserClaim.UserID, page, limit, startedAtGTE, startedAtLT)
	if err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) getCurrentUserWorkoutAnalytics(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	startedAtGTE, startedAtLT, err := a.workoutAnalyticsDateRange(r)
	if err != nil {
		handler.BadRequest(w, "start_date and end_date must be valid YYYY-MM-DD values, and end_date must be on or after start_date")
		return
	}

	response, err := a.App.WorkoutService.GetWorkoutAnalytics(r.Context(), ctx.UserClaim.UserID, ctx.UserClaim.UserID, startedAtGTE, startedAtLT)
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

func (a *API) listCurrentUserPersonalRecords(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	page, limit, err := a.paginationParams(r)
	if err != nil {
		handler.BadRequest(w, "page and limit must be positive integers")
		return
	}

	response, err := a.App.WorkoutService.ListCurrentUserPersonalRecords(r.Context(), ctx.UserClaim.UserID, page, limit)
	if err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) getUserWorkoutAnalytics(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	startedAtGTE, startedAtLT, err := a.workoutAnalyticsDateRange(r)
	if err != nil {
		handler.BadRequest(w, "start_date and end_date must be valid YYYY-MM-DD values, and end_date must be on or after start_date")
		return
	}

	userID := pathID(r)
	response, err := a.App.WorkoutService.GetWorkoutAnalytics(r.Context(), ctx.UserClaim.UserID, userID, startedAtGTE, startedAtLT)
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
