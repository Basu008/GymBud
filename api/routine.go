package api

import (
	"errors"
	"net/http"

	approutine "github.com/Basu008/GymBud/app/routine"
	"github.com/Basu008/GymBud/schema"
	"github.com/Basu008/GymBud/server/handler"
	"github.com/gorilla/mux"
)

func (a *API) listRoutines(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	response, err := a.App.RoutineService.ListRoutines(r.Context(), ctx.UserClaim.UserID)
	if err != nil {
		handler.InternalServerError(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) getRoutineByID(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	routineID := mux.Vars(r)["id"]
	response, err := a.App.RoutineService.GetRoutineByID(r.Context(), ctx.UserClaim.UserID, routineID)
	if err != nil {
		if errors.Is(err, approutine.ErrRoutineNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) createRoutine(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.CreateRoutineBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &body) {
		return
	}

	response, err := a.App.RoutineService.CreateRoutine(r.Context(), ctx.UserClaim.UserID, &body)
	if err != nil {
		switch {
		case errors.Is(err, approutine.ErrRoutineLimitReached):
			handler.Conflict(w, err.Error())
		case errors.Is(err, approutine.ErrExerciseNotFound):
			handler.NotFound(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.Created(w, response)
}

func (a *API) copyRoutine(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	routineID := mux.Vars(r)["id"]
	response, err := a.App.RoutineService.CopyRoutine(r.Context(), ctx.UserClaim.UserID, routineID)
	if err != nil {
		switch {
		case errors.Is(err, approutine.ErrRoutineNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, approutine.ErrExerciseNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, approutine.ErrRoutineLimitReached):
			handler.Conflict(w, err.Error())
		case errors.Is(err, approutine.ErrRoutineCopyNotAllowed):
			handler.Forbidden(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.Created(w, response)
}

func (a *API) updateRoutine(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.UpdateRoutineBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	routineID := mux.Vars(r)["id"]
	response, err := a.App.RoutineService.UpdateRoutine(r.Context(), ctx.UserClaim.UserID, routineID, &body)
	if err != nil {
		switch {
		case errors.Is(err, approutine.ErrRoutineNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, approutine.ErrExerciseNotFound):
			handler.NotFound(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) deleteRoutine(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	routineID := mux.Vars(r)["id"]
	response, err := a.App.RoutineService.DeleteRoutine(r.Context(), ctx.UserClaim.UserID, routineID)
	if err != nil {
		if errors.Is(err, approutine.ErrRoutineNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}
