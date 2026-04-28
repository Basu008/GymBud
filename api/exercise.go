package api

import (
	"errors"
	"net/http"
	"strings"

	appexercise "github.com/Basu008/GymBud/app/exercise"
	"github.com/Basu008/GymBud/schema"
	"github.com/Basu008/GymBud/server/handler"
)

func (a *API) listExercises(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	category := r.URL.Query().Get("category")

	if strings.TrimSpace(category) != "" {
		if err := validateCategoryValue(category); err != nil {
			handler.BadRequest(w, err.Error())
			return
		}
	}

	response, err := a.App.ExerciseService.ListExercises(r.Context(), &name, &category)
	if err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
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
	if err := validateExerciseInput(body.Category, body.Equipment, body.MovementMode); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	response, err := a.App.ExerciseService.CreateExercise(r.Context(), &body)
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
	if err := validatePartialExerciseInput(body.Category, body.Equipment, body.MovementMode); err != nil {
		handler.BadRequest(w, err.Error())
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

func validateExerciseInput(category, equipment string, movementMode *string) error {
	categoryOptions := map[string]struct{}{
		"chest":     {},
		"back":      {},
		"triceps":   {},
		"biceps":    {},
		"shoulders": {},
		"legs":      {},
		"abs":       {},
		"forearms":  {},
	}
	equipmentOptions := map[string]struct{}{
		"dumbbell":    {},
		"barbell":     {},
		"cables":      {},
		"machine":     {},
		"body weight": {},
	}
	category = strings.TrimSpace(strings.ToLower(category))
	equipment = strings.TrimSpace(strings.ToLower(equipment))
	mode := ""
	if movementMode != nil {
		mode = strings.TrimSpace(strings.ToLower(*movementMode))
	}
	if category != "" {
		if _, ok := categoryOptions[category]; !ok {
			return errors.New("category must be one of: Chest, Back, Triceps, Biceps, Shoulders, Legs, Abs, Forearms")
		}
	}

	if mode != "" && mode != "unilateral" && mode != "bilateral" {
		return errors.New("movement_mode must be unilateral or bilateral")
	}
	if equipment != "" {
		if _, ok := equipmentOptions[equipment]; !ok {
			return errors.New("equipment must be one of: Dumbbell, Barbell, Cables, Machine, Body Weight")
		}
	}

	if equipment == "dumbbell" || equipment == "cables" || equipment == "machine" {
		if mode == "" {
			return errors.New("movement_mode is required when equipment is Dumbbell, Cables, or Machine")
		}
	}

	return nil
}

func validatePartialExerciseInput(category, equipment, movementMode *string) error {
	if category != nil {
		value := strings.TrimSpace(*category)
		if value != "" {
			if err := validateCategoryValue(value); err != nil {
				return err
			}
		}
	}

	if equipment != nil {
		value := strings.TrimSpace(*equipment)
		if value != "" {
			if err := validateEquipmentValue(value); err != nil {
				return err
			}
		}
	}

	if movementMode != nil {
		mode := strings.TrimSpace(strings.ToLower(*movementMode))
		if mode != "" && mode != "unilateral" && mode != "bilateral" {
			return errors.New("movement_mode must be unilateral or bilateral")
		}
	}

	return nil
}

func validateCategoryValue(value string) error {
	valid := map[string]struct{}{
		"chest":     {},
		"back":      {},
		"triceps":   {},
		"biceps":    {},
		"shoulders": {},
		"legs":      {},
		"abs":       {},
		"forearms":  {},
	}

	if _, ok := valid[strings.TrimSpace(strings.ToLower(value))]; !ok {
		return errors.New("category must be one of: Chest, Back, Triceps, Biceps, Shoulders, Legs, Abs, Forearms")
	}

	return nil
}

func validateEquipmentValue(value string) error {
	valid := map[string]struct{}{
		"dumbbell":    {},
		"barbell":     {},
		"cables":      {},
		"machine":     {},
		"body weight": {},
	}

	if _, ok := valid[strings.TrimSpace(strings.ToLower(value))]; !ok {
		return errors.New("equipment must be one of: Dumbbell, Barbell, Cables, Machine, Body Weight")
	}

	return nil
}
