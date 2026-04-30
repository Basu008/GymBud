package api

import (
	"errors"
	"net/http"

	appexercise "github.com/Basu008/GymBud/app/exercise"
	"github.com/Basu008/GymBud/schema"
	"github.com/Basu008/GymBud/server/handler"
)

func (a *API) listExercises(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	category := r.URL.Query().Get("category")

	response, err := a.App.ExerciseService.ListExercises(r.Context(), ctx.UserClaim.UserID, &name, &category)
	if err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) listExerciseCategories(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	handler.OK(w, a.App.ExerciseService.ListExerciseCategories())
}

func (a *API) listExerciseMuscles(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	handler.OK(w, a.App.ExerciseService.ListExerciseMuscles())
}

func (a *API) listExerciseEquipments(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	handler.OK(w, a.App.ExerciseService.ListExerciseEquipments())
}

func (a *API) listExerciseDifficulty(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	handler.OK(w, a.App.ExerciseService.ListExerciseDifficulty())
}

func (a *API) getExerciseByID(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	exerciseID := pathID(r)
	response, err := a.App.ExerciseService.GetExerciseByID(r.Context(), exerciseID)
	if err != nil {
		switch {
		case errors.Is(err, appexercise.ErrExerciseNotFound):
			handler.NotFound(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) createExercise(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.CreateExerciseBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &body) {
		return
	}

	response, err := a.App.ExerciseService.CreateExercise(r.Context(), ctx.UserClaim.UserID, &body)
	if err != nil {
		if errors.Is(err, appexercise.ErrExerciseNameAlreadyExists) {
			handler.Conflict(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.Created(w, response)
}

func (a *API) createExercises(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.CreateExercisesBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &body) {
		return
	}

	response, err := a.App.ExerciseService.CreateExercises(r.Context(), ctx.UserClaim.UserID, &body)
	if err != nil {
		if errors.Is(err, appexercise.ErrExerciseNameAlreadyExists) {
			handler.Conflict(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.Created(w, response)
}

func (a *API) updateExercise(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.UpdateExerciseBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &body) {
		return
	}

	exerciseID := pathID(r)
	response, err := a.App.ExerciseService.UpdateExercise(r.Context(), exerciseID, &body)
	if err != nil {
		switch {
		case errors.Is(err, appexercise.ErrExerciseNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, appexercise.ErrExerciseNameAlreadyExists):
			handler.Conflict(w, err.Error())
		case errors.Is(err, appexercise.ErrExerciseManagedByAdmin):
			handler.Forbidden(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) deleteExercise(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	exerciseID := pathID(r)
	response, err := a.App.ExerciseService.DeleteExercise(r.Context(), exerciseID)
	if err != nil {
		switch {
		case errors.Is(err, appexercise.ErrExerciseNotFound):
			handler.NotFound(w, err.Error())
		case errors.Is(err, appexercise.ErrExerciseManagedByAdmin):
			handler.Forbidden(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}
