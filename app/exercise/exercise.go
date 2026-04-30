package exercise

import (
	"context"
	"errors"
	"strings"
	"unicode"

	modelexercise "github.com/Basu008/GymBud/model/exercise"
	"github.com/Basu008/GymBud/schema"
	"github.com/google/uuid"
)

var ErrExerciseNotFound = errors.New("exercise not found")
var ErrExerciseNameAlreadyExists = errors.New("exercise already exists for this equipment")
var ErrExerciseManagedByAdmin = errors.New("admin-created exercises cannot be updated or deleted")

func (s *Service) ListExercises(ctx context.Context, userID string, nameRegex, category *string) (*schema.ExercisesResponse, error) {
	filter := &modelexercise.ListFilter{
		NameRegex: normalizeOptionalString(nameRegex),
		Category:  normalizeOptionalLowerString(category),
		UserID:    strings.TrimSpace(userID),
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

func (s *Service) ListExerciseCategories() *schema.ExerciseCategoriesResponse {
	return &schema.ExerciseCategoriesResponse{
		ExerciseCategories: cloneStrings(s.config.AdditionalConfig.ExerciseCategories),
	}
}

func (s *Service) ListExerciseMuscles() *schema.ExerciseMusclesResponse {
	return &schema.ExerciseMusclesResponse{
		ExerciseMuscles: cloneStrings(s.config.AdditionalConfig.ExerciseMuscles),
	}
}

func (s *Service) ListExerciseEquipments() *schema.ExerciseEquipmentsResponse {
	return &schema.ExerciseEquipmentsResponse{
		ExerciseEquipments: cloneStrings(s.config.AdditionalConfig.ExerciseEquipments),
	}
}

func (s *Service) ListExerciseDifficulty() *schema.ExerciseDifficultyResponse {
	return &schema.ExerciseDifficultyResponse{
		ExerciseDifficulty: cloneStrings(s.config.AdditionalConfig.ExerciseDifficulty),
	}
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

func (s *Service) CreateExercise(ctx context.Context, userID string, body *schema.CreateExerciseBody) (*schema.ExerciseResponse, error) {
	exercise, err := s.createExercise(ctx, userID, body)
	if err != nil {
		return nil, err
	}

	return &schema.ExerciseResponse{Exercise: toExercisePayload(exercise)}, nil
}

func (s *Service) CreateExercises(ctx context.Context, userID string, body *schema.CreateExercisesBody) (*schema.ExercisesResponse, error) {
	if body == nil || len(body.Exercises) == 0 {
		return nil, errors.New("at least one exercise is required")
	}

	payload := make([]*schema.ExercisePayload, 0, len(body.Exercises))
	for i := range body.Exercises {
		if body.Exercises[i].IsAdmin == nil {
			isAdmin := true
			body.Exercises[i].IsAdmin = &isAdmin
		}

		exercise, err := s.createExercise(ctx, userID, &body.Exercises[i])
		if err != nil {
			return nil, err
		}
		payload = append(payload, toExercisePayload(exercise))
	}

	return &schema.ExercisesResponse{Exercises: payload}, nil
}

func (s *Service) createExercise(ctx context.Context, userID string, body *schema.CreateExerciseBody) (*modelexercise.Exercise, error) {
	isAdmin := false
	if body.IsAdmin != nil {
		isAdmin = *body.IsAdmin
	}
	userID = strings.TrimSpace(userID)
	var ownerID *string
	if !isAdmin {
		if _, err := uuid.Parse(userID); err != nil {
			return nil, err
		}
		ownerID = &userID
	}

	exercise := &modelexercise.Exercise{
		ID:               uuid.NewString(),
		Name:             strings.TrimSpace(body.Name),
		Slug:             slugify(body.Name),
		Category:         strings.TrimSpace(body.Category),
		UserID:           ownerID,
		Equipment:        strings.TrimSpace(body.Equipment),
		PrimaryMuscle:    strings.TrimSpace(body.PrimaryMuscle),
		SecondaryMuscles: normalizeStringSlice(body.SecondaryMuscles),
		Difficulty:       strings.TrimSpace(body.Difficulty),
		MovementMode:     normalizeMovementMode(body.MovementMode),
		IsMadeByAdmin:    isAdmin,
		IsActive:         true,
	}
	if exercise.Slug == "" {
		return nil, errors.New("name must contain at least one letter or number")
	}

	if err := s.validateExerciseValues(exercise.Category, exercise.Equipment, exercise.PrimaryMuscle, exercise.SecondaryMuscles, exercise.Difficulty, exercise.MovementMode); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, exercise); err != nil {
		if errors.Is(err, modelexercise.ErrExerciseNameAlreadyExists) {
			return nil, ErrExerciseNameAlreadyExists
		}
		return nil, err
	}

	return exercise, nil
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
	nextPrimaryMuscle := currentExercise.PrimaryMuscle
	nextSecondaryMuscles := currentExercise.SecondaryMuscles
	nextDifficulty := currentExercise.Difficulty
	nextMovementMode := currentExercise.MovementMode

	if body.Name != nil {
		value := strings.TrimSpace(*body.Name)
		slug := slugify(value)
		if slug == "" {
			return nil, errors.New("name must contain at least one letter or number")
		}
		updateInput.NameSet = true
		updateInput.Name = &value
		updateInput.SlugSet = true
		updateInput.Slug = &slug
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
	if body.PrimaryMuscle != nil {
		value := strings.TrimSpace(*body.PrimaryMuscle)
		updateInput.PrimaryMuscleSet = true
		updateInput.PrimaryMuscle = &value
		nextPrimaryMuscle = value
	}
	if body.SecondaryMuscles != nil {
		value := normalizeStringSlice(*body.SecondaryMuscles)
		updateInput.SecondaryMusclesSet = true
		updateInput.SecondaryMuscles = value
		nextSecondaryMuscles = value
	}
	if body.Difficulty != nil {
		value := strings.TrimSpace(*body.Difficulty)
		updateInput.DifficultySet = true
		updateInput.Difficulty = &value
		nextDifficulty = value
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

	if !updateInput.NameSet && !updateInput.CategorySet && !updateInput.EquipmentSet && !updateInput.PrimaryMuscleSet && !updateInput.SecondaryMusclesSet && !updateInput.DifficultySet && !updateInput.MovementModeSet && !updateInput.IsMadeByAdminSet && !updateInput.IsActiveSet {
		return nil, errors.New("at least one updatable field is required")
	}

	if err := s.validateExerciseValues(nextCategory, nextEquipment, nextPrimaryMuscle, nextSecondaryMuscles, nextDifficulty, nextMovementMode); err != nil {
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

func normalizeStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
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
		ID:               exercise.ID,
		Name:             exercise.Name,
		Slug:             exercise.Slug,
		Category:         exercise.Category,
		UserID:           exercise.UserID,
		Equipment:        exercise.Equipment,
		PrimaryMuscle:    exercise.PrimaryMuscle,
		SecondaryMuscles: exercise.SecondaryMuscles,
		Difficulty:       exercise.Difficulty,
		MovementMode:     exercise.MovementMode,
		IsActive:         exercise.IsActive,
		CreatedAt:        exercise.CreatedAt,
		UpdatedAt:        exercise.UpdatedAt,
	}
}

func (s *Service) validateExerciseValues(category, equipment, primaryMuscle string, secondaryMuscles []string, difficulty string, movementMode *string) error {
	category = strings.TrimSpace(category)
	equipment = strings.TrimSpace(equipment)
	primaryMuscle = strings.TrimSpace(primaryMuscle)
	difficulty = strings.TrimSpace(difficulty)
	mode := ""
	if movementMode != nil {
		mode = strings.TrimSpace(strings.ToLower(*movementMode))
	}

	additionalConfig := s.config.AdditionalConfig
	if !containsFold(additionalConfig.ExerciseCategories, category) {
		return errors.New("category must be one of: " + strings.Join(additionalConfig.ExerciseCategories, ", "))
	}
	if !containsFold(additionalConfig.ExerciseEquipments, equipment) {
		return errors.New("equipment must be one of: " + strings.Join(additionalConfig.ExerciseEquipments, ", "))
	}
	if !containsFold(additionalConfig.ExerciseMuscles, primaryMuscle) {
		return errors.New("primary_muscle must be one of: " + strings.Join(additionalConfig.ExerciseMuscles, ", "))
	}
	for _, muscle := range secondaryMuscles {
		if !containsFold(additionalConfig.ExerciseMuscles, muscle) {
			return errors.New("secondary_muscles must contain only: " + strings.Join(additionalConfig.ExerciseMuscles, ", "))
		}
	}
	if !containsFold(additionalConfig.ExerciseDifficulty, difficulty) {
		return errors.New("difficulty must be one of: " + strings.Join(additionalConfig.ExerciseDifficulty, ", "))
	}
	if mode != "" && mode != "unilateral" && mode != "bilateral" {
		return errors.New("movement_mode must be unilateral or bilateral")
	}
	if requiresMovementMode(equipment) && mode == "" {
		return errors.New("movement_mode is required when equipment is Dumbbell, Cable, or Machine")
	}

	return nil
}

func slugify(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var builder strings.Builder
	lastWasHyphen := false

	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastWasHyphen = false
		case !lastWasHyphen:
			builder.WriteByte('-')
			lastWasHyphen = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func containsFold(options []string, value string) bool {
	for _, option := range options {
		if strings.EqualFold(strings.TrimSpace(option), strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}

func requiresMovementMode(equipment string) bool {
	equipment = strings.TrimSpace(equipment)
	return strings.EqualFold(equipment, "Dumbbell") ||
		strings.EqualFold(equipment, "Cable") ||
		strings.EqualFold(equipment, "Machine")
}

func cloneStrings(values []string) []string {
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}
