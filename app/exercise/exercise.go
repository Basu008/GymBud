package exercise

import (
	"context"
	"errors"
	"strings"

	modelexercise "github.com/Basu008/GymBud/model/exercise"
	"github.com/Basu008/GymBud/schema"
	"github.com/google/uuid"
)

var ErrExerciseNotFound = errors.New("exercise not found")
var ErrExerciseNameAlreadyExists = errors.New("exercise already exists for this equipment")
var ErrExerciseManagedByAdmin = errors.New("admin-created exercises cannot be updated or deleted")

func (s *Service) ListExercises(ctx context.Context, nameRegex, category *string) (*schema.ExercisesResponse, error) {
	filter := &modelexercise.ListFilter{
		NameRegex: normalizeOptionalString(nameRegex),
		Category:  normalizeOptionalLowerString(category),
	}

	exercises, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	payload := make([]*schema.ExercisePayload, 0, len(exercises))
	for _, exercise := range exercises {
		payload = append(payload, toExercisePayload(exercise))
	}

	return &schema.ExercisesResponse{Exercises: payload}, nil
}

func (s *Service) GetExerciseByID(ctx context.Context, exerciseID string) (*schema.ExerciseResponse, error) {
	exerciseID = strings.TrimSpace(exerciseID)
	exercise, err := s.repo.GetByID(ctx, exerciseID)
	if err != nil {
		if errors.Is(err, modelexercise.ErrExerciseNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}

	return &schema.ExerciseResponse{Exercise: toExercisePayload(exercise)}, nil
}

func (s *Service) CreateExercise(ctx context.Context, body *schema.CreateExerciseBody) (*schema.ExerciseResponse, error) {
	isAdmin := false
	if body.IsAdmin != nil {
		isAdmin = *body.IsAdmin
	}

	exercise := &modelexercise.Exercise{
		ID:            uuid.NewString(),
		Name:          strings.TrimSpace(body.Name),
		Category:      strings.TrimSpace(body.Category),
		Equipment:     strings.TrimSpace(body.Equipment),
		MovementMode:  normalizeMovementMode(body.MovementMode),
		IsMadeByAdmin: isAdmin,
		IsActive:      true,
	}

	if err := s.repo.Create(ctx, exercise); err != nil {
		if errors.Is(err, modelexercise.ErrExerciseNameAlreadyExists) {
			return nil, ErrExerciseNameAlreadyExists
		}
		return nil, err
	}

	return &schema.ExerciseResponse{Exercise: toExercisePayload(exercise)}, nil
}

func (s *Service) UpdateExercise(ctx context.Context, exerciseID string, body *schema.UpdateExerciseBody) (*schema.ExerciseResponse, error) {
	exerciseID = strings.TrimSpace(exerciseID)
	currentExercise, err := s.repo.GetByID(ctx, exerciseID)
	if err != nil {
		if errors.Is(err, modelexercise.ErrExerciseNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}

	updateInput := &modelexercise.UpdateInput{}
	nextCategory := currentExercise.Category
	nextEquipment := currentExercise.Equipment
	nextMovementMode := currentExercise.MovementMode

	if body.Name != nil {
		value := strings.TrimSpace(*body.Name)
		updateInput.NameSet = true
		updateInput.Name = &value
	}
	if body.Category != nil {
		value := strings.TrimSpace(*body.Category)
		updateInput.CategorySet = true
		updateInput.Category = &value
		nextCategory = value
	}
	if body.Equipment != nil {
		value := strings.TrimSpace(*body.Equipment)
		updateInput.EquipmentSet = true
		updateInput.Equipment = &value
		nextEquipment = value
	}
	if body.MovementMode != nil {
		updateInput.MovementModeSet = true
		updateInput.MovementMode = normalizeMovementMode(body.MovementMode)
		nextMovementMode = updateInput.MovementMode
	}
	if body.IsAdmin != nil {
		updateInput.IsMadeByAdminSet = true
		updateInput.IsMadeByAdmin = body.IsAdmin
	}
	if body.IsActive != nil {
		updateInput.IsActiveSet = true
		updateInput.IsActive = body.IsActive
	}

	if !updateInput.NameSet && !updateInput.CategorySet && !updateInput.EquipmentSet && !updateInput.MovementModeSet && !updateInput.IsMadeByAdminSet && !updateInput.IsActiveSet {
		return nil, errors.New("at least one updatable field is required")
	}

	if err := validateExerciseValues(nextCategory, nextEquipment, nextMovementMode); err != nil {
		return nil, err
	}

	exercise, err := s.repo.UpdateByID(ctx, exerciseID, updateInput)
	if err != nil {
		switch {
		case errors.Is(err, modelexercise.ErrExerciseNotFound):
			return nil, ErrExerciseNotFound
		case errors.Is(err, modelexercise.ErrExerciseNameAlreadyExists):
			return nil, ErrExerciseNameAlreadyExists
		case errors.Is(err, modelexercise.ErrExerciseManagedByAdmin):
			return nil, ErrExerciseManagedByAdmin
		default:
			return nil, err
		}
	}

	return &schema.ExerciseResponse{Exercise: toExercisePayload(exercise)}, nil
}

func (s *Service) DeleteExercise(ctx context.Context, exerciseID string) (*schema.DeleteExerciseResponse, error) {
	exerciseID = strings.TrimSpace(exerciseID)
	if _, err := uuid.Parse(exerciseID); err != nil {
		return nil, err
	}

	if err := s.repo.DeleteByID(ctx, exerciseID); err != nil {
		switch {
		case errors.Is(err, modelexercise.ErrExerciseNotFound):
			return nil, ErrExerciseNotFound
		case errors.Is(err, modelexercise.ErrExerciseManagedByAdmin):
			return nil, ErrExerciseManagedByAdmin
		default:
			return nil, err
		}
	}

	return &schema.DeleteExerciseResponse{DeletedID: exerciseID}, nil
}

func normalizeMovementMode(mode *string) *string {
	if mode == nil {
		return nil
	}

	value := strings.TrimSpace(*mode)
	if value == "" {
		return nil
	}

	return &value
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeOptionalLowerString(value *string) *string {
	normalized := normalizeOptionalString(value)
	if normalized == nil {
		return nil
	}

	lower := strings.ToLower(*normalized)
	return &lower
}

func toExercisePayload(exercise *modelexercise.Exercise) *schema.ExercisePayload {
	return &schema.ExercisePayload{
		ID:           exercise.ID,
		Name:         exercise.Name,
		Category:     exercise.Category,
		Equipment:    exercise.Equipment,
		MovementMode: exercise.MovementMode,
		IsActive:     exercise.IsActive,
		CreatedAt:    exercise.CreatedAt,
		UpdatedAt:    exercise.UpdatedAt,
	}
}

func validateExerciseValues(category, equipment string, movementMode *string) error {
	category = strings.TrimSpace(strings.ToLower(category))
	equipment = strings.TrimSpace(strings.ToLower(equipment))
	mode := ""
	if movementMode != nil {
		mode = strings.TrimSpace(strings.ToLower(*movementMode))
	}

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

	if _, ok := categoryOptions[category]; !ok {
		return errors.New("category must be one of: Chest, Back, Triceps, Biceps, Shoulders, Legs, Abs, Forearms")
	}
	if _, ok := equipmentOptions[equipment]; !ok {
		return errors.New("equipment must be one of: Dumbbell, Barbell, Cables, Machine, Body Weight")
	}
	if mode != "" && mode != "unilateral" && mode != "bilateral" {
		return errors.New("movement_mode must be unilateral or bilateral")
	}
	if (equipment == "dumbbell" || equipment == "cables" || equipment == "machine") && mode == "" {
		return errors.New("movement_mode is required when equipment is Dumbbell, Cables, or Machine")
	}

	return nil
}
