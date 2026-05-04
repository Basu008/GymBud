package workout

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	modelroutine "github.com/Basu008/GymBud/model/routine"
	modelsocial "github.com/Basu008/GymBud/model/social"
	modeluser "github.com/Basu008/GymBud/model/user"
	modelworkout "github.com/Basu008/GymBud/model/workout"
	"github.com/Basu008/GymBud/schema"
	"github.com/google/uuid"
)

var ErrRoutineNotFound = errors.New("routine not found")
var ErrRoutineExerciseNotFound = errors.New("one or more exercises do not exist in the routine")
var ErrWorkoutNotFound = errors.New("workout not found")
var ErrUserNotFound = errors.New("user not found")
var ErrWorkoutAccessDenied = errors.New("you are not allowed to view this user's workouts")
var ErrWorkoutDeleteNotAllowed = errors.New("you are not allowed to delete this workout")

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
		record.WorkoutID = workout.ID
		if err := s.repo.CreatePersonalRecord(ctx, record); err != nil {
			return nil, err
		}
	}

	return &schema.WorkoutResponse{Workout: toWorkoutPayload(workout, nil, true, workout.Stats.PRCount > 0, nil)}, nil
}

func (s *Service) ListUserWorkouts(ctx context.Context, viewerUserID, targetUserID string, page, limit int, startedAtGTE, startedAtLT *time.Time) (*schema.WorkoutsResponse, error) {
	viewerUserID = strings.TrimSpace(viewerUserID)
	targetUserID = strings.TrimSpace(targetUserID)

	if _, err := uuid.Parse(viewerUserID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(targetUserID); err != nil {
		return nil, err
	}

	visibility, err := s.resolveWorkoutVisibility(ctx, viewerUserID, targetUserID)
	if err != nil {
		return nil, err
	}

	offset := int64((page - 1) * limit)
	workouts, total, err := s.repo.ListByUserID(ctx, &modelworkout.ListFilter{
		UserID:       targetUserID,
		Visibility:   visibility,
		StartedAtGTE: startedAtGTE,
		StartedAtLT:  startedAtLT,
		Offset:       offset,
		Limit:        int64(limit),
	})
	if err != nil {
		return nil, err
	}

	currentPRWorkoutIDs := map[string]bool{}
	likeSummaries := map[string]*modelsocial.WorkoutLikeSummary{}
	if len(workouts) > 0 {
		workoutIDs := make([]string, 0, len(workouts))
		for _, workout := range workouts {
			workoutIDs = append(workoutIDs, workout.ID)
		}

		currentPRWorkoutIDs, err = s.repo.GetCurrentPRWorkoutIDs(ctx, targetUserID, workoutIDs)
		if err != nil {
			return nil, err
		}
		likeSummaries, err = s.socialRepo.GetWorkoutLikeSummaries(ctx, viewerUserID, workoutIDs)
		if err != nil {
			return nil, err
		}
	}

	payloads := make([]*schema.WorkoutPayload, 0, len(workouts))
	for _, workout := range workouts {
		payloads = append(payloads, toWorkoutPayload(workout, nil, viewerUserID == workout.UserID, currentPRWorkoutIDs[workout.ID], likeSummaries[workout.ID]))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &schema.WorkoutsResponse{
		Workouts: payloads,
		Pagination: schema.PaginationPayload{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *Service) GetWorkoutAnalytics(ctx context.Context, viewerUserID, targetUserID string, startedAtGTE, startedAtLT *time.Time) (*schema.WorkoutAnalyticsResponse, error) {
	viewerUserID = strings.TrimSpace(viewerUserID)
	targetUserID = strings.TrimSpace(targetUserID)

	if _, err := uuid.Parse(viewerUserID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(targetUserID); err != nil {
		return nil, err
	}

	visibility, err := s.resolveWorkoutVisibility(ctx, viewerUserID, targetUserID)
	if err != nil {
		return nil, err
	}

	workouts, err := s.repo.ListAllByUserID(ctx, &modelworkout.ListFilter{
		UserID:       targetUserID,
		Visibility:   visibility,
		StartedAtGTE: startedAtGTE,
		StartedAtLT:  startedAtLT,
	})
	if err != nil {
		return nil, err
	}

	stats := schema.WorkoutAnalyticsStatsPayload{
		WorkoutsCount: len(workouts),
	}
	for _, workout := range workouts {
		stats.TotalVolume += workout.Stats.TotalVolume
		stats.TotalSets += workout.Stats.TotalSets
		stats.TotalReps += workout.Stats.TotalReps
		stats.PRCount += workout.Stats.PRCount
	}
	stats.TotalVolume = math.Round(stats.TotalVolume*100) / 100

	return &schema.WorkoutAnalyticsResponse{
		UserID: targetUserID,
		Stats:  stats,
	}, nil
}

func (s *Service) ListCurrentUserPersonalRecords(ctx context.Context, userID string, page, limit int) (*schema.PersonalRecordsResponse, error) {
	userID = strings.TrimSpace(userID)

	if _, err := uuid.Parse(userID); err != nil {
		return nil, err
	}

	offset := int64((page - 1) * limit)
	records, total, err := s.repo.ListCurrentPersonalRecords(ctx, userID, offset, int64(limit))
	if err != nil {
		return nil, err
	}

	payloads := make([]*schema.PersonalRecordPayload, 0, len(records))
	for _, record := range records {
		payloads = append(payloads, toPersonalRecordPayload(record))
	}

	return &schema.PersonalRecordsResponse{
		PersonalRecords: payloads,
		Pagination:      schema.NewPaginationPayload(page, limit, total),
	}, nil
}

func (s *Service) ListFollowingWorkouts(ctx context.Context, viewerUserID string, page, limit int, startedAtGTE, startedAtLT *time.Time) (*schema.WorkoutsResponse, error) {
	viewerUserID = strings.TrimSpace(viewerUserID)

	if _, err := uuid.Parse(viewerUserID); err != nil {
		return nil, err
	}

	followingIDs, err := s.socialRepo.ListFollowingIDs(ctx, viewerUserID)
	if err != nil {
		return nil, err
	}

	publicVisibility := "all"
	workouts := make([]*modelworkout.Workout, 0)
	var total int64
	if len(followingIDs) > 0 {
		offset := int64((page - 1) * limit)
		workouts, total, err = s.repo.ListFeedByUserIDs(ctx, &modelworkout.ListFilter{
			UserIDs:      followingIDs,
			Visibility:   &publicVisibility,
			StartedAtGTE: startedAtGTE,
			StartedAtLT:  startedAtLT,
			Offset:       offset,
			Limit:        int64(limit),
		})
		if err != nil {
			return nil, err
		}
	}

	currentPRWorkoutIDs := map[string]bool{}
	likeSummaries := map[string]*modelsocial.WorkoutLikeSummary{}
	if len(workouts) > 0 {
		workoutIDs := make([]string, 0, len(workouts))
		workoutIDsByUser := make(map[string][]string)
		for _, workout := range workouts {
			workoutIDs = append(workoutIDs, workout.ID)
			workoutIDsByUser[workout.UserID] = append(workoutIDsByUser[workout.UserID], workout.ID)
		}

		for ownerUserID, ownerWorkoutIDs := range workoutIDsByUser {
			ownerPRWorkoutIDs, err := s.repo.GetCurrentPRWorkoutIDs(ctx, ownerUserID, ownerWorkoutIDs)
			if err != nil {
				return nil, err
			}
			for workoutID, hasPR := range ownerPRWorkoutIDs {
				currentPRWorkoutIDs[workoutID] = hasPR
			}
		}

		likeSummaries, err = s.socialRepo.GetWorkoutLikeSummaries(ctx, viewerUserID, workoutIDs)
		if err != nil {
			return nil, err
		}
	}

	ownerByUserID := make(map[string]*schema.WorkoutUserPayload)
	for _, workout := range workouts {
		if _, exists := ownerByUserID[workout.UserID]; exists {
			continue
		}

		owner, err := s.userRepo.GetByID(ctx, workout.UserID)
		if err != nil {
			if errors.Is(err, modeluser.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		ownerByUserID[workout.UserID] = toWorkoutUserPayload(owner)
	}

	payloads := make([]*schema.WorkoutPayload, 0, len(workouts))
	for _, workout := range workouts {
		payloads = append(payloads, toWorkoutPayload(workout, ownerByUserID[workout.UserID], false, currentPRWorkoutIDs[workout.ID], likeSummaries[workout.ID]))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &schema.WorkoutsResponse{
		Workouts: payloads,
		Pagination: schema.PaginationPayload{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *Service) GetWorkoutByID(ctx context.Context, viewerUserID, workoutID string) (*schema.WorkoutResponse, error) {
	viewerUserID = strings.TrimSpace(viewerUserID)
	workoutID = strings.TrimSpace(workoutID)

	if _, err := uuid.Parse(viewerUserID); err != nil {
		return nil, err
	}
	if workoutID == "" {
		return nil, errors.New("workout id is required")
	}

	workout, err := s.repo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, modelworkout.ErrWorkoutNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}

	if viewerUserID != workout.UserID {
		owner, err := s.userRepo.GetByID(ctx, workout.UserID)
		if err != nil {
			if errors.Is(err, modeluser.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}

		if owner.IsPrivate {
			isFollowing, err := s.socialRepo.IsFollowing(ctx, viewerUserID, workout.UserID)
			if err != nil {
				return nil, err
			}
			if !isFollowing {
				return nil, ErrWorkoutAccessDenied
			}
		}

		if workout.Visibility != "all" {
			return nil, ErrWorkoutAccessDenied
		}
	}

	containsPR, err := s.repo.GetCurrentPRWorkoutIDs(ctx, workout.UserID, []string{workout.ID})
	if err != nil {
		return nil, err
	}
	likeSummaries, err := s.socialRepo.GetWorkoutLikeSummaries(ctx, viewerUserID, []string{workout.ID})
	if err != nil {
		return nil, err
	}

	return &schema.WorkoutResponse{
		Workout: toWorkoutPayload(workout, nil, viewerUserID == workout.UserID, containsPR[workout.ID], likeSummaries[workout.ID]),
	}, nil
}

func (s *Service) GetLatestWorkoutByRoutineID(ctx context.Context, userID, routineID string) (*schema.WorkoutResponse, error) {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)

	if _, err := uuid.Parse(userID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(routineID); err != nil {
		return nil, err
	}

	if _, err := s.routineRepo.GetByID(ctx, userID, routineID); err != nil {
		if errors.Is(err, modelroutine.ErrRoutineNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}

	workout, err := s.repo.GetLatestByRoutineID(ctx, userID, routineID)
	if err != nil {
		if errors.Is(err, modelworkout.ErrWorkoutNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}

	containsPR, err := s.repo.GetCurrentPRWorkoutIDs(ctx, userID, []string{workout.ID})
	if err != nil {
		return nil, err
	}
	likeSummaries, err := s.socialRepo.GetWorkoutLikeSummaries(ctx, userID, []string{workout.ID})
	if err != nil {
		return nil, err
	}

	return &schema.WorkoutResponse{
		Workout: toWorkoutPayload(workout, nil, true, containsPR[workout.ID], likeSummaries[workout.ID]),
	}, nil
}

func (s *Service) DeleteWorkout(ctx context.Context, userID, workoutID string) (*schema.DeleteWorkoutResponse, error) {
	userID = strings.TrimSpace(userID)
	workoutID = strings.TrimSpace(workoutID)

	if _, err := uuid.Parse(userID); err != nil {
		return nil, err
	}
	if workoutID == "" {
		return nil, errors.New("workout id is required")
	}

	workout, err := s.repo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, modelworkout.ErrWorkoutNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}

	if workout.UserID != userID {
		return nil, ErrWorkoutDeleteNotAllowed
	}

	if err := s.repo.Delete(ctx, workoutID); err != nil {
		if errors.Is(err, modelworkout.ErrWorkoutNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	if err := s.socialRepo.DeleteWorkoutLikes(ctx, workoutID); err != nil {
		return nil, err
	}

	return &schema.DeleteWorkoutResponse{DeletedID: workoutID}, nil
}

func (s *Service) LikeWorkout(ctx context.Context, userID, workoutID string) (*schema.WorkoutResponse, error) {
	userID = strings.TrimSpace(userID)
	workoutID = strings.TrimSpace(workoutID)

	if _, err := uuid.Parse(userID); err != nil {
		return nil, err
	}
	if workoutID == "" {
		return nil, errors.New("workout id is required")
	}

	workout, err := s.repo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, modelworkout.ErrWorkoutNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	if err := s.socialRepo.LikeWorkout(ctx, workoutID, userID); err != nil {
		return nil, err
	}
	likeSummaries, err := s.socialRepo.GetWorkoutLikeSummaries(ctx, userID, []string{workoutID})
	if err != nil {
		return nil, err
	}

	return &schema.WorkoutResponse{Workout: toWorkoutPayload(workout, nil, userID == workout.UserID, workout.Stats.PRCount > 0, likeSummaries[workoutID])}, nil
}

func (s *Service) UnlikeWorkout(ctx context.Context, userID, workoutID string) (*schema.WorkoutResponse, error) {
	userID = strings.TrimSpace(userID)
	workoutID = strings.TrimSpace(workoutID)

	if _, err := uuid.Parse(userID); err != nil {
		return nil, err
	}
	if workoutID == "" {
		return nil, errors.New("workout id is required")
	}

	workout, err := s.repo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, modelworkout.ErrWorkoutNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	if err := s.socialRepo.UnlikeWorkout(ctx, workoutID, userID); err != nil {
		return nil, err
	}
	likeSummaries, err := s.socialRepo.GetWorkoutLikeSummaries(ctx, userID, []string{workoutID})
	if err != nil {
		return nil, err
	}

	return &schema.WorkoutResponse{Workout: toWorkoutPayload(workout, nil, userID == workout.UserID, workout.Stats.PRCount > 0, likeSummaries[workoutID])}, nil
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
	bestSetIndex := -1

	for idx, input := range inputs {
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

		sets = append(sets, &modelworkout.WorkoutExerciseSet{
			SetNumber:       input.SetNumber,
			PlannedMinReps:  routineSet.MinReps,
			PlannedMaxReps:  routineSet.MaxReps,
			PlannedWeightKG: routineSet.TargetWeightKG,
			ActualReps:      input.ActualReps,
			ActualWeightKG:  input.ActualWeightKG,
			PRFlags:         modelworkout.PRFlags{},
		})
		if bestSetIndex == -1 || isBetterWorkoutSet(input, inputs[bestSetIndex]) {
			bestSetIndex = idx
		}

		stats.TotalSets++
		stats.TotalReps += input.ActualReps
		stats.TotalVolume += float64(input.ActualReps) * input.ActualWeightKG
	}

	if bestSetIndex >= 0 {
		bestSet := inputs[bestSetIndex]
		prFlags, _, createdRecord := evaluatePR(workout, routineExercise, bestSet.ActualWeightKG, bestSet.ActualReps, now, currentRecord)
		sets[bestSetIndex].PRFlags = prFlags
		if createdRecord != nil {
			stats.PRCount = 1
			return sets, stats, []*modelworkout.PersonalRecord{createdRecord}, nil
		}
	}

	return sets, stats, nil, nil
}

func isBetterWorkoutSet(candidate, currentBest schema.CreateWorkoutSetInput) bool {
	candidate1RM := estimateOneRM(candidate.ActualWeightKG, candidate.ActualReps)
	currentBest1RM := estimateOneRM(currentBest.ActualWeightKG, currentBest.ActualReps)

	if candidate1RM != currentBest1RM {
		return candidate1RM > currentBest1RM
	}
	if candidate.ActualWeightKG != currentBest.ActualWeightKG {
		return candidate.ActualWeightKG > currentBest.ActualWeightKG
	}
	if candidate.ActualReps != currentBest.ActualReps {
		return candidate.ActualReps > currentBest.ActualReps
	}
	return candidate.SetNumber < currentBest.SetNumber
}

func evaluatePR(workout *modelworkout.Workout, routineExercise *modelroutine.RoutineExercise, actualWeight float64, actualReps int, now time.Time, current *modelworkout.PersonalRecord) (modelworkout.PRFlags, *modelworkout.PersonalRecord, *modelworkout.PersonalRecord) {
	current1RM := estimateOneRM(actualWeight, actualReps)
	flags := modelworkout.PRFlags{}

	if current == nil {
		record := &modelworkout.PersonalRecord{
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

func toWorkoutPayload(workout *modelworkout.Workout, owner *schema.WorkoutUserPayload, includeSensitiveFields bool, hasPR bool, likeSummary *modelsocial.WorkoutLikeSummary) *schema.WorkoutPayload {
	exercises := make([]*schema.WorkoutExercisePayload, 0, len(workout.Exercises))
	for _, exercise := range workout.Exercises {
		sets := make([]*schema.WorkoutSetPayload, 0, len(exercise.Sets))
		for _, set := range exercise.Sets {
			payloadSet := &schema.WorkoutSetPayload{
				SetNumber:      set.SetNumber,
				ActualReps:     set.ActualReps,
				ActualWeightKG: set.ActualWeightKG,
				PRFlags: schema.WorkoutPRFlagsPayload{
					WeightPR:       set.PRFlags.WeightPR,
					RepPR:          set.PRFlags.RepPR,
					Estimated1RMPR: set.PRFlags.Estimated1RMPR,
				},
			}
			if includeSensitiveFields {
				payloadSet.PlannedMinReps = intPointer(set.PlannedMinReps)
				payloadSet.PlannedMaxReps = intPointer(set.PlannedMaxReps)
				payloadSet.PlannedWeightKG = set.PlannedWeightKG
			}
			sets = append(sets, payloadSet)
		}

		exercises = append(exercises, &schema.WorkoutExercisePayload{
			ExerciseID:   exercise.ExerciseID,
			ExerciseName: exercise.ExerciseName,
			OrderIndex:   exercise.OrderIndex,
			Sets:         sets,
		})
	}

	likesCount := 0
	likedByMe := false
	if likeSummary != nil {
		likesCount = likeSummary.LikesCount
		likedByMe = likeSummary.LikedByMe
	}

	return &schema.WorkoutPayload{
		ID:          workout.ID,
		UserID:      workout.UserID,
		User:        owner,
		RoutineID:   workout.RoutineID,
		Title:       workout.Title,
		StartedAt:   workout.StartedAt,
		EndedAt:     workout.EndedAt,
		DurationSec: workout.DurationSec,
		Visibility:  workout.Visibility,
		Notes:       notesForViewer(workout.Notes, includeSensitiveFields),
		HasPR:       hasPR,
		LikesCount:  likesCount,
		LikedByMe:   likedByMe,
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

func toWorkoutUserPayload(user *modeluser.User) *schema.WorkoutUserPayload {
	if user == nil {
		return nil
	}

	return &schema.WorkoutUserPayload{
		ID:              user.ID.String(),
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		ProfileImageURL: user.ProfileImageURL,
	}
}

func intPointer(value int) *int {
	return &value
}

func notesForViewer(notes *string, includeNotes bool) *string {
	if !includeNotes {
		return nil
	}
	return notes
}

func toPersonalRecordPayload(record *modelworkout.PersonalRecord) *schema.PersonalRecordPayload {
	if record == nil {
		return nil
	}

	return &schema.PersonalRecordPayload{
		ID:           record.ID,
		UserID:       record.UserID,
		ExerciseID:   record.ExerciseID,
		ExerciseName: record.ExerciseName,
		BestWeightKG: record.BestWeightKG,
		BestReps:     record.BestReps,
		Estimated1RM: record.Estimated1RM,
		WorkoutID:    record.WorkoutID,
		UpdatedAt:    record.UpdatedAt,
	}
}

func (s *Service) resolveWorkoutVisibility(ctx context.Context, viewerUserID, targetUserID string) (*string, error) {
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	visibility := (*string)(nil)
	if viewerUserID != targetUserID {
		publicVisibility := "all"
		visibility = &publicVisibility

		if targetUser.IsPrivate {
			isFollowing, err := s.socialRepo.IsFollowing(ctx, viewerUserID, targetUserID)
			if err != nil {
				return nil, err
			}
			if !isFollowing {
				return nil, ErrWorkoutAccessDenied
			}
		}
	}

	return visibility, nil
}
