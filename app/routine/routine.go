package routine

import (
	"context"
	"errors"
	"strings"

	modelroutine "github.com/Basu008/GymBud/model/routine"
	"github.com/Basu008/GymBud/schema"
	"github.com/google/uuid"
)

const maxRoutinesPerUser = 9

var ErrRoutineNotFound = errors.New("routine not found")
var ErrRoutineLimitReached = errors.New("users can have at most 9 routines")
var ErrExerciseNotFound = errors.New("one or more exercises not found")
var ErrRoutineCopyNotAllowed = errors.New("you can copy this routine only if you follow its owner")

func (s *Service) ListRoutines(ctx context.Context, userID string) (*schema.RoutinesResponse, error) {
	routines, err := s.repo.ListByUserID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	payload := make([]*schema.RoutinePayload, 0, len(routines))
	for _, routine := range routines {
		payload = append(payload, toRoutinePayload(routine))
	}

	return &schema.RoutinesResponse{Routines: payload}, nil
}

func (s *Service) GetRoutineByID(ctx context.Context, userID, routineID string) (*schema.RoutineResponse, error) {
	routine, err := s.repo.GetByID(ctx, strings.TrimSpace(userID), strings.TrimSpace(routineID))
	if err != nil {
		if errors.Is(err, modelroutine.ErrRoutineNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}

	return &schema.RoutineResponse{Routine: toRoutinePayload(routine)}, nil
}

func (s *Service) CreateRoutine(ctx context.Context, userID string, body *schema.CreateRoutineBody) (*schema.RoutineResponse, error) {
	userID = strings.TrimSpace(userID)
	if err := s.ensureCanCreateRoutine(ctx, userID); err != nil {
		return nil, err
	}

	routine, err := s.buildRoutine(userID, strings.TrimSpace(body.Name), body.Exercises)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, routine)
	if err != nil {
		if errors.Is(err, modelroutine.ErrRoutineExerciseNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}

	return &schema.RoutineResponse{Routine: toRoutinePayload(created)}, nil
}

func (s *Service) CopyRoutine(ctx context.Context, currentUserID, routineID string) (*schema.RoutineResponse, error) {
	currentUserID = strings.TrimSpace(currentUserID)
	routineID = strings.TrimSpace(routineID)

	if err := s.ensureCanCreateRoutine(ctx, currentUserID); err != nil {
		return nil, err
	}

	source, err := s.repo.GetByRoutineID(ctx, routineID)
	if err != nil {
		if errors.Is(err, modelroutine.ErrRoutineNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}

	if source.UserID != currentUserID {
		isFollowing, err := s.socialRepo.IsFollowing(ctx, currentUserID, source.UserID)
		if err != nil {
			return nil, err
		}
		if !isFollowing {
			return nil, ErrRoutineCopyNotAllowed
		}
	}

	clonedRoutine, err := s.cloneRoutine(currentUserID, source)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, clonedRoutine)
	if err != nil {
		if errors.Is(err, modelroutine.ErrRoutineExerciseNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}

	return &schema.RoutineResponse{Routine: toRoutinePayload(created)}, nil
}

func (s *Service) UpdateRoutine(ctx context.Context, userID, routineID string, body *schema.UpdateRoutineBody) (*schema.RoutineResponse, error) {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)

	current, err := s.repo.GetByID(ctx, userID, routineID)
	if err != nil {
		if errors.Is(err, modelroutine.ErrRoutineNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}

	changed := false
	if body.Name != nil {
		current.Name = strings.TrimSpace(*body.Name)
		changed = true
	}
	if body.Exercises != nil {
		exercises, err := s.buildRoutineExercises(*body.Exercises)
		if err != nil {
			return nil, err
		}
		current.Exercises = exercises
		changed = true
	}
	if !changed {
		return nil, errors.New("at least one updatable field is required")
	}
	if strings.TrimSpace(current.Name) == "" {
		return nil, errors.New("name is required")
	}

	updated, err := s.repo.ReplaceByID(ctx, userID, current)
	if err != nil {
		switch {
		case errors.Is(err, modelroutine.ErrRoutineNotFound):
			return nil, ErrRoutineNotFound
		case errors.Is(err, modelroutine.ErrRoutineExerciseNotFound):
			return nil, ErrExerciseNotFound
		default:
			return nil, err
		}
	}

	return &schema.RoutineResponse{Routine: toRoutinePayload(updated)}, nil
}

func (s *Service) DeleteRoutine(ctx context.Context, userID, routineID string) (*schema.DeleteRoutineResponse, error) {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)

	if err := s.repo.DeleteByID(ctx, userID, routineID); err != nil {
		if errors.Is(err, modelroutine.ErrRoutineNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}

	return &schema.DeleteRoutineResponse{DeletedID: routineID}, nil
}

func (s *Service) ensureCanCreateRoutine(ctx context.Context, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return err
	}

	count, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if count >= maxRoutinesPerUser {
		return ErrRoutineLimitReached
	}

	return nil
}

func (s *Service) buildRoutine(userID, name string, exerciseInputs []schema.RoutineExerciseInput) (*modelroutine.Routine, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	exercises, err := s.buildRoutineExercises(exerciseInputs)
	if err != nil {
		return nil, err
	}

	return &modelroutine.Routine{
		ID:        uuid.NewString(),
		UserID:    userID,
		Name:      name,
		Exercises: exercises,
	}, nil
}

func (s *Service) buildRoutineExercises(inputs []schema.RoutineExerciseInput) ([]*modelroutine.RoutineExercise, error) {
	if len(inputs) == 0 {
		return nil, errors.New("at least one exercise is required")
	}

	exercises := make([]*modelroutine.RoutineExercise, 0, len(inputs))
	orderIndexes := make(map[int]struct{}, len(inputs))
	exerciseIDs := make(map[string]struct{}, len(inputs))

	for _, input := range inputs {
		exerciseID := strings.TrimSpace(input.ExerciseID)
		if exerciseID == "" {
			return nil, errors.New("exercise_id is required")
		}
		if _, err := uuid.Parse(exerciseID); err != nil {
			return nil, err
		}
		if input.OrderIndex <= 0 {
			return nil, errors.New("order_index must be greater than 0")
		}
		if _, exists := orderIndexes[input.OrderIndex]; exists {
			return nil, errors.New("order_index must be unique within a routine")
		}
		orderIndexes[input.OrderIndex] = struct{}{}

		if _, exists := exerciseIDs[exerciseID]; exists {
			return nil, errors.New("exercise_id must be unique within a routine")
		}
		exerciseIDs[exerciseID] = struct{}{}

		sets, err := buildRoutineExerciseSets(input.Sets)
		if err != nil {
			return nil, err
		}

		exercises = append(exercises, &modelroutine.RoutineExercise{
			ID:         uuid.NewString(),
			ExerciseID: exerciseID,
			OrderIndex: input.OrderIndex,
			Notes:      normalizeOptionalText(input.Notes),
			Sets:       sets,
		})
	}

	for expected := 1; expected <= len(inputs); expected++ {
		if _, ok := orderIndexes[expected]; !ok {
			return nil, errors.New("order_index values must be continuous starting from 1")
		}
	}

	return exercises, nil
}

func buildRoutineExerciseSets(inputs []schema.RoutineSetInput) ([]*modelroutine.RoutineExerciseSet, error) {
	if len(inputs) == 0 {
		return nil, errors.New("each routine exercise must include at least one set")
	}

	sets := make([]*modelroutine.RoutineExerciseSet, 0, len(inputs))
	setNumbers := make(map[int]struct{}, len(inputs))

	for _, input := range inputs {
		if input.SetNumber <= 0 {
			return nil, errors.New("set_number must be greater than 0")
		}
		if _, exists := setNumbers[input.SetNumber]; exists {
			return nil, errors.New("set_number must be unique within a routine exercise")
		}
		setNumbers[input.SetNumber] = struct{}{}

		if input.MinReps <= 0 {
			return nil, errors.New("min_reps must be greater than 0")
		}
		if input.MaxReps < input.MinReps {
			return nil, errors.New("max_reps must be greater than or equal to min_reps")
		}
		if input.TargetWeightKG != nil && *input.TargetWeightKG < 0 {
			return nil, errors.New("target_weight_kg must be greater than or equal to 0")
		}

		sets = append(sets, &modelroutine.RoutineExerciseSet{
			ID:             uuid.NewString(),
			SetNumber:      input.SetNumber,
			MinReps:        input.MinReps,
			MaxReps:        input.MaxReps,
			TargetWeightKG: input.TargetWeightKG,
		})
	}

	for expected := 1; expected <= len(inputs); expected++ {
		if _, ok := setNumbers[expected]; !ok {
			return nil, errors.New("set_number values must be continuous starting from 1")
		}
	}

	return sets, nil
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func (s *Service) cloneRoutine(userID string, source *modelroutine.Routine) (*modelroutine.Routine, error) {
	exercises := make([]*modelroutine.RoutineExercise, 0, len(source.Exercises))
	for _, sourceExercise := range source.Exercises {
		sets := make([]*modelroutine.RoutineExerciseSet, 0, len(sourceExercise.Sets))
		for _, sourceSet := range sourceExercise.Sets {
			var targetWeight *float64
			if sourceSet.TargetWeightKG != nil {
				value := *sourceSet.TargetWeightKG
				targetWeight = &value
			}

			sets = append(sets, &modelroutine.RoutineExerciseSet{
				ID:             uuid.NewString(),
				SetNumber:      sourceSet.SetNumber,
				MinReps:        sourceSet.MinReps,
				MaxReps:        sourceSet.MaxReps,
				TargetWeightKG: targetWeight,
			})
		}

		exercises = append(exercises, &modelroutine.RoutineExercise{
			ID:         uuid.NewString(),
			ExerciseID: sourceExercise.ExerciseID,
			OrderIndex: sourceExercise.OrderIndex,
			Notes:      normalizeOptionalText(sourceExercise.Notes),
			Sets:       sets,
		})
	}

	return &modelroutine.Routine{
		ID:        uuid.NewString(),
		UserID:    userID,
		Name:      source.Name,
		Exercises: exercises,
	}, nil
}

func toRoutinePayload(routine *modelroutine.Routine) *schema.RoutinePayload {
	exercises := make([]*schema.RoutineExercisePayload, 0, len(routine.Exercises))
	for _, exercise := range routine.Exercises {
		sets := make([]*schema.RoutineSetPayload, 0, len(exercise.Sets))
		for _, set := range exercise.Sets {
			sets = append(sets, &schema.RoutineSetPayload{
				ID:                set.ID,
				RoutineExerciseID: set.RoutineExerciseID,
				SetNumber:         set.SetNumber,
				MinReps:           set.MinReps,
				MaxReps:           set.MaxReps,
				TargetWeightKG:    set.TargetWeightKG,
				CreatedAt:         set.CreatedAt,
				UpdatedAt:         set.UpdatedAt,
			})
		}

		var exercisePayload *schema.ExercisePayload
		if exercise.Exercise != nil {
			exercisePayload = &schema.ExercisePayload{
				ID:               exercise.Exercise.ID,
				Name:             exercise.Exercise.Name,
				Slug:             exercise.Exercise.Slug,
				Category:         exercise.Exercise.Category,
				UserID:           exercise.Exercise.UserID,
				Equipment:        exercise.Exercise.Equipment,
				PrimaryMuscle:    exercise.Exercise.PrimaryMuscle,
				SecondaryMuscles: exercise.Exercise.SecondaryMuscles,
				Difficulty:       exercise.Exercise.Difficulty,
				MovementMode:     exercise.Exercise.MovementMode,
				IsActive:         exercise.Exercise.IsActive,
				CreatedAt:        exercise.Exercise.CreatedAt,
				UpdatedAt:        exercise.Exercise.UpdatedAt,
			}
		}

		exercises = append(exercises, &schema.RoutineExercisePayload{
			ID:         exercise.ID,
			RoutineID:  exercise.RoutineID,
			ExerciseID: exercise.ExerciseID,
			OrderIndex: exercise.OrderIndex,
			Notes:      exercise.Notes,
			Exercise:   exercisePayload,
			Sets:       sets,
			CreatedAt:  exercise.CreatedAt,
			UpdatedAt:  exercise.UpdatedAt,
		})
	}

	return &schema.RoutinePayload{
		ID:        routine.ID,
		UserID:    routine.UserID,
		Name:      routine.Name,
		Exercises: exercises,
		CreatedAt: routine.CreatedAt,
		UpdatedAt: routine.UpdatedAt,
	}
}
