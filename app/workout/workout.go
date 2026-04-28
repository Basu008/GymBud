package workout

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	modelroutine "github.com/Basu008/GymBud/model/routine"
	modelworkout "github.com/Basu008/GymBud/model/workout"
	"github.com/Basu008/GymBud/schema"
	"github.com/google/uuid"
)

var ErrRoutineNotFound = errors.New("routine not found")
var ErrRoutineExerciseNotFound = errors.New("one or more exercises do not exist in the routine")

func (s *Service) CreateWorkout(ctx context.Context, userID string, body *schema.CreateWorkoutBody) (*schema.WorkoutResponse, error) {
	userID = strings.TrimSpace(userID)
	routineID := strings.TrimSpace(body.RoutineID)

	if _, err := uuid.Parse(userID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(routineID); err != nil {
		return nil, err
	}

	startedAt := body.StartTime.UTC()
	endedAt := body.EndTime.UTC()
	if startedAt.IsZero() {
		return nil, errors.New("start_time is required")
	}
	if endedAt.IsZero() {
		return nil, errors.New("end_time is required")
	}
	if !endedAt.After(startedAt) {
		return nil, errors.New("end_time must be after start_time")
	}

	visibility := strings.ToLower(strings.TrimSpace(body.Visibility))
	if visibility != "all" && visibility != "private" {
		return nil, errors.New("visibility must be one of: all, private")
	}

	if len(body.Exercises) == 0 {
		return nil, errors.New("at least one exercise is required")
	}

	routine, err := s.routineRepo.GetByID(ctx, userID, routineID)
	if err != nil {
		if errors.Is(err, modelroutine.ErrRoutineNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}

	now := time.Now().UTC()
	workout := &modelworkout.Workout{
		UserID:      userID,
		RoutineID:   routine.ID,
		Title:       routine.Name,
		StartedAt:   startedAt,
		EndedAt:     endedAt,
		DurationSec: int(endedAt.Sub(startedAt).Seconds()),
		Visibility:  visibility,
		Notes:       normalizeOptionalText(body.Notes),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	workoutExercises, stats, personalRecords, err := s.buildWorkoutExercises(ctx, workout, body.Exercises, routine.Exercises, now)
	if err != nil {
		return nil, err
	}
	workout.Exercises = workoutExercises
	workout.Stats = stats

	if err := s.repo.Create(ctx, workout); err != nil {
		return nil, err
	}
	for _, record := range personalRecords {
		if err := s.repo.CreatePersonalRecord(ctx, record); err != nil {
			return nil, err
		}
	}

	return &schema.WorkoutResponse{Workout: toWorkoutPayload(workout)}, nil
}

func (s *Service) buildWorkoutExercises(ctx context.Context, workout *modelworkout.Workout, inputs []schema.CreateWorkoutExerciseInput, routineExercises []*modelroutine.RoutineExercise, now time.Time) ([]*modelworkout.WorkoutExercise, modelworkout.WorkoutStats, []*modelworkout.PersonalRecord, error) {
	routineExerciseByID := make(map[string]*modelroutine.RoutineExercise, len(routineExercises))
	for _, routineExercise := range routineExercises {
		routineExerciseByID[routineExercise.ExerciseID] = routineExercise
	}

	seenExerciseIDs := make(map[string]struct{}, len(inputs))
	workoutExercises := make([]*modelworkout.WorkoutExercise, 0, len(inputs))
	stats := modelworkout.WorkoutStats{}
	personalRecords := make([]*modelworkout.PersonalRecord, 0)

	for _, input := range inputs {
		exerciseID := strings.TrimSpace(input.ExerciseID)
		if exerciseID == "" {
			return nil, stats, nil, errors.New("exercise_id is required")
		}
		if _, err := uuid.Parse(exerciseID); err != nil {
			return nil, stats, nil, err
		}
		if _, exists := seenExerciseIDs[exerciseID]; exists {
			return nil, stats, nil, errors.New("exercise_id must be unique within a workout")
		}
		seenExerciseIDs[exerciseID] = struct{}{}

		routineExercise, ok := routineExerciseByID[exerciseID]
		if !ok {
			return nil, stats, nil, ErrRoutineExerciseNotFound
		}
		if routineExercise.Exercise == nil {
			return nil, stats, nil, errors.New("routine exercise metadata is missing")
		}
		if len(input.Sets) == 0 {
			return nil, stats, nil, errors.New("each exercise must include at least one set")
		}

		currentRecord, err := s.repo.GetLatestPersonalRecord(ctx, workout.UserID, exerciseID)
		if err != nil {
			return nil, stats, nil, err
		}

		sets, exerciseStats, createdRecords, err := buildWorkoutSets(workout, routineExercise, input.Sets, currentRecord, now)
		if err != nil {
			return nil, stats, nil, err
		}

		stats.TotalSets += exerciseStats.TotalSets
		stats.TotalReps += exerciseStats.TotalReps
		stats.TotalVolume += exerciseStats.TotalVolume
		stats.PRCount += exerciseStats.PRCount
		if len(createdRecords) > 0 {
			personalRecords = append(personalRecords, createdRecords...)
		}

		workoutExercises = append(workoutExercises, &modelworkout.WorkoutExercise{
			ExerciseID:   routineExercise.ExerciseID,
			ExerciseName: routineExercise.Exercise.Name,
			OrderIndex:   routineExercise.OrderIndex,
			Sets:         sets,
		})
	}

	stats.TotalVolume = math.Round(stats.TotalVolume*100) / 100

	return workoutExercises, stats, personalRecords, nil
}

func buildWorkoutSets(workout *modelworkout.Workout, routineExercise *modelroutine.RoutineExercise, inputs []schema.CreateWorkoutSetInput, currentRecord *modelworkout.PersonalRecord, now time.Time) ([]*modelworkout.WorkoutExerciseSet, modelworkout.WorkoutStats, []*modelworkout.PersonalRecord, error) {
	routineSetByNumber := make(map[int]*modelroutine.RoutineExerciseSet, len(routineExercise.Sets))
	for _, routineSet := range routineExercise.Sets {
		routineSetByNumber[routineSet.SetNumber] = routineSet
	}

	seenSetNumbers := make(map[int]struct{}, len(inputs))
	sets := make([]*modelworkout.WorkoutExerciseSet, 0, len(inputs))
	stats := modelworkout.WorkoutStats{}
	workingRecord := clonePersonalRecord(currentRecord)
	createdRecords := make([]*modelworkout.PersonalRecord, 0)
	exerciseHadPR := false

	for _, input := range inputs {
		if input.SetNumber <= 0 {
			return nil, stats, nil, errors.New("set_number must be greater than 0")
		}
		if _, exists := seenSetNumbers[input.SetNumber]; exists {
			return nil, stats, nil, errors.New("set_number must be unique within an exercise")
		}
		seenSetNumbers[input.SetNumber] = struct{}{}

		routineSet, ok := routineSetByNumber[input.SetNumber]
		if !ok {
			return nil, stats, nil, errors.New("one or more sets do not exist in the routine exercise")
		}
		if input.ActualReps <= 0 {
			return nil, stats, nil, errors.New("actual_reps must be greater than 0")
		}
		if input.ActualWeightKG < 0 {
			return nil, stats, nil, errors.New("actual_weight_kg must be greater than or equal to 0")
		}

		prFlags, nextRecord, createdRecord := evaluatePR(workout, routineExercise, input.ActualWeightKG, input.ActualReps, now, workingRecord)
		if prFlags.WeightPR || prFlags.RepPR || prFlags.Estimated1RMPR {
			exerciseHadPR = true
		}
		workingRecord = nextRecord
		if createdRecord != nil {
			createdRecords = append(createdRecords, createdRecord)
		}

		sets = append(sets, &modelworkout.WorkoutExerciseSet{
			SetNumber:       input.SetNumber,
			PlannedMinReps:  routineSet.MinReps,
			PlannedMaxReps:  routineSet.MaxReps,
			PlannedWeightKG: routineSet.TargetWeightKG,
			ActualReps:      input.ActualReps,
			ActualWeightKG:  input.ActualWeightKG,
			PRFlags:         prFlags,
		})

		stats.TotalSets++
		stats.TotalReps += input.ActualReps
		stats.TotalVolume += float64(input.ActualReps) * input.ActualWeightKG
	}

	if exerciseHadPR {
		stats.PRCount = 1
	}

	return sets, stats, createdRecords, nil
}

func evaluatePR(workout *modelworkout.Workout, routineExercise *modelroutine.RoutineExercise, actualWeight float64, actualReps int, now time.Time, current *modelworkout.PersonalRecord) (modelworkout.PRFlags, *modelworkout.PersonalRecord, *modelworkout.PersonalRecord) {
	current1RM := estimateOneRM(actualWeight, actualReps)
	flags := modelworkout.PRFlags{}

	if current == nil {
		record := &modelworkout.PersonalRecord{
			ID:           personalRecordID(),
			UserID:       workout.UserID,
			ExerciseID:   routineExercise.ExerciseID,
			ExerciseName: routineExercise.Exercise.Name,
			BestWeightKG: actualWeight,
			BestReps:     actualReps,
			Estimated1RM: current1RM,
			WorkoutID:    workout.ID,
			UpdatedAt:    now,
		}
		return modelworkout.PRFlags{
			WeightPR:       true,
			RepPR:          true,
			Estimated1RMPR: true,
		}, clonePersonalRecord(record), record
	}

	next := clonePersonalRecord(current)
	if actualWeight > current.BestWeightKG {
		flags.WeightPR = true
		next.BestWeightKG = actualWeight
		next.BestReps = actualReps
	}
	if actualWeight == current.BestWeightKG && actualReps > current.BestReps {
		flags.RepPR = true
		next.BestReps = actualReps
	}
	if current1RM > current.Estimated1RM {
		flags.Estimated1RMPR = true
		next.Estimated1RM = current1RM
	}

	if flags.WeightPR || flags.RepPR || flags.Estimated1RMPR {
		next.ExerciseName = routineExercise.Exercise.Name
		next.WorkoutID = workout.ID
		next.UpdatedAt = now
		record := &modelworkout.PersonalRecord{
			ID:           personalRecordID(),
			UserID:       workout.UserID,
			ExerciseID:   routineExercise.ExerciseID,
			ExerciseName: next.ExerciseName,
			BestWeightKG: next.BestWeightKG,
			BestReps:     next.BestReps,
			Estimated1RM: next.Estimated1RM,
			WorkoutID:    next.WorkoutID,
			UpdatedAt:    next.UpdatedAt,
		}
		return flags, next, record
	}

	return flags, current, nil
}

func toWorkoutPayload(workout *modelworkout.Workout) *schema.WorkoutPayload {
	exercises := make([]*schema.WorkoutExercisePayload, 0, len(workout.Exercises))
	for _, exercise := range workout.Exercises {
		sets := make([]*schema.WorkoutSetPayload, 0, len(exercise.Sets))
		for _, set := range exercise.Sets {
			sets = append(sets, &schema.WorkoutSetPayload{
				SetNumber:       set.SetNumber,
				PlannedMinReps:  set.PlannedMinReps,
				PlannedMaxReps:  set.PlannedMaxReps,
				PlannedWeightKG: set.PlannedWeightKG,
				ActualReps:      set.ActualReps,
				ActualWeightKG:  set.ActualWeightKG,
				PRFlags: schema.WorkoutPRFlagsPayload{
					WeightPR:       set.PRFlags.WeightPR,
					RepPR:          set.PRFlags.RepPR,
					Estimated1RMPR: set.PRFlags.Estimated1RMPR,
				},
			})
		}

		exercises = append(exercises, &schema.WorkoutExercisePayload{
			ExerciseID:   exercise.ExerciseID,
			ExerciseName: exercise.ExerciseName,
			OrderIndex:   exercise.OrderIndex,
			Sets:         sets,
		})
	}

	return &schema.WorkoutPayload{
		ID:          workout.ID,
		UserID:      workout.UserID,
		RoutineID:   workout.RoutineID,
		Title:       workout.Title,
		StartedAt:   workout.StartedAt,
		EndedAt:     workout.EndedAt,
		DurationSec: workout.DurationSec,
		Visibility:  workout.Visibility,
		Notes:       workout.Notes,
		Exercises:   exercises,
		Stats: schema.WorkoutStatsPayload{
			TotalSets:   workout.Stats.TotalSets,
			TotalReps:   workout.Stats.TotalReps,
			TotalVolume: workout.Stats.TotalVolume,
			PRCount:     workout.Stats.PRCount,
		},
		CreatedAt: workout.CreatedAt,
		UpdatedAt: workout.UpdatedAt,
	}
}

func estimateOneRM(weight float64, reps int) float64 {
	value := weight * (1 + float64(reps)/30)
	return math.Round(value*100) / 100
}

func clonePersonalRecord(record *modelworkout.PersonalRecord) *modelworkout.PersonalRecord {
	if record == nil {
		return nil
	}

	copy := *record
	return &copy
}

func personalRecordID() string {
	return "pr_" + uuid.NewString()
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
