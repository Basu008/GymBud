package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	modelworkout "github.com/Basu008/GymBud/model/workout"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type WorkoutRepo struct {
	collection               *mongodriver.Collection
	personalRecordCollection *mongodriver.Collection
}

type workoutDocument struct {
	ID          bson.ObjectID                   `bson:"_id,omitempty"`
	UserID      string                          `bson:"user_id"`
	RoutineID   string                          `bson:"routine_id"`
	Title       string                          `bson:"title"`
	StartedAt   time.Time                       `bson:"started_at"`
	EndedAt     time.Time                       `bson:"ended_at"`
	DurationSec int                             `bson:"duration_sec"`
	Visibility  string                          `bson:"visibility"`
	Notes       *string                         `bson:"notes,omitempty"`
	Exercises   []*modelworkout.WorkoutExercise `bson:"exercises"`
	Stats       modelworkout.WorkoutStats       `bson:"stats"`
	CreatedAt   time.Time                       `bson:"created_at"`
	UpdatedAt   time.Time                       `bson:"updated_at"`
}

type personalRecordDocument struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	UserID       string        `bson:"user_id"`
	ExerciseID   string        `bson:"exercise_id"`
	ExerciseName string        `bson:"exercise_name"`
	BestWeightKG float64       `bson:"best_weight_kg"`
	BestReps     int           `bson:"best_reps"`
	Estimated1RM float64       `bson:"estimated_1rm"`
	WorkoutID    string        `bson:"workout_id"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

func NewWorkoutRepo(db *mongodriver.Database) (*WorkoutRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("mongo database is required")
	}

	return &WorkoutRepo{
		collection:               db.Collection("workouts"),
		personalRecordCollection: db.Collection("personal_records"),
	}, nil
}

func (r *WorkoutRepo) Create(ctx context.Context, workout *modelworkout.Workout) error {
	result, err := r.collection.InsertOne(ctx, workoutDocumentFromModel(workout))
	if err != nil {
		return fmt.Errorf("insert workout: %w", err)
	}
	objectID, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return fmt.Errorf("insert workout: unexpected inserted id type %T", result.InsertedID)
	}
	workout.ID = objectID.Hex()

	return nil
}

func (r *WorkoutRepo) GetByID(ctx context.Context, workoutID string) (*modelworkout.Workout, error) {
	objectID, err := bson.ObjectIDFromHex(workoutID)
	if err != nil {
		return nil, modelworkout.ErrWorkoutNotFound
	}

	var workout workoutDocument
	if err := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&workout); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return nil, modelworkout.ErrWorkoutNotFound
		}
		return nil, fmt.Errorf("find workout: %w", err)
	}

	return workoutModelFromDocument(&workout), nil
}

func (r *WorkoutRepo) GetLatestByRoutineID(ctx context.Context, userID, routineID string) (*modelworkout.Workout, error) {
	var workout workoutDocument
	err := r.collection.FindOne(
		ctx,
		bson.M{"user_id": userID, "routine_id": routineID},
		options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}),
	).Decode(&workout)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return nil, modelworkout.ErrWorkoutNotFound
		}
		return nil, fmt.Errorf("find latest workout by routine: %w", err)
	}

	return workoutModelFromDocument(&workout), nil
}

func (r *WorkoutRepo) Delete(ctx context.Context, workoutID string) error {
	objectID, err := bson.ObjectIDFromHex(workoutID)
	if err != nil {
		return modelworkout.ErrWorkoutNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("delete workout: %w", err)
	}
	if result.DeletedCount == 0 {
		return modelworkout.ErrWorkoutNotFound
	}

	if _, err := r.personalRecordCollection.DeleteMany(ctx, bson.M{"workout_id": workoutID}); err != nil {
		return fmt.Errorf("delete workout personal records: %w", err)
	}

	return nil
}

func (r *WorkoutRepo) ListByUserID(ctx context.Context, filter *modelworkout.ListFilter) ([]*modelworkout.Workout, int64, error) {
	query := buildWorkoutListQuery(filter)

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("count workouts: %w", err)
	}

	cursor, err := r.collection.Find(
		ctx,
		query,
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
			SetSkip(filter.Offset).
			SetLimit(filter.Limit),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list workouts: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*workoutDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("decode workouts: %w", err)
	}

	workouts := make([]*modelworkout.Workout, 0, len(docs))
	for _, doc := range docs {
		workouts = append(workouts, workoutModelFromDocument(doc))
	}

	return workouts, total, nil
}

func (r *WorkoutRepo) ListAllByUserID(ctx context.Context, filter *modelworkout.ListFilter) ([]*modelworkout.Workout, error) {
	query := buildWorkoutListQuery(filter)

	cursor, err := r.collection.Find(
		ctx,
		query,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("list all workouts: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*workoutDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode all workouts: %w", err)
	}

	workouts := make([]*modelworkout.Workout, 0, len(docs))
	for _, doc := range docs {
		workouts = append(workouts, workoutModelFromDocument(doc))
	}

	return workouts, nil
}

func (r *WorkoutRepo) ListFeedByUserIDs(ctx context.Context, filter *modelworkout.ListFilter) ([]*modelworkout.Workout, int64, error) {
	query := buildWorkoutListQuery(filter)

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("count feed workouts: %w", err)
	}

	cursor, err := r.collection.Find(
		ctx,
		query,
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
			SetSkip(filter.Offset).
			SetLimit(filter.Limit),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list feed workouts: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*workoutDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("decode feed workouts: %w", err)
	}

	workouts := make([]*modelworkout.Workout, 0, len(docs))
	for _, doc := range docs {
		workouts = append(workouts, workoutModelFromDocument(doc))
	}

	return workouts, total, nil
}

func (r *WorkoutRepo) GetCurrentPRWorkoutIDs(ctx context.Context, userID string, workoutIDs []string) (map[string]bool, error) {
	if len(workoutIDs) == 0 {
		return map[string]bool{}, nil
	}

	pipeline := mongodriver.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$sort", Value: bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: -1}}}},
		{{Key: "$group", Value: bson.M{
			"_id":        "$exercise_id",
			"workout_id": bson.M{"$first": "$workout_id"},
		}}},
		{{Key: "$match", Value: bson.M{"workout_id": bson.M{"$in": workoutIDs}}}},
	}

	cursor, err := r.personalRecordCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate current pr workouts: %w", err)
	}
	defer cursor.Close(ctx)

	type currentPRWorkoutResult struct {
		WorkoutID string `bson:"workout_id"`
	}

	var results []currentPRWorkoutResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode current pr workouts: %w", err)
	}

	workoutIDSet := make(map[string]bool, len(results))
	for _, result := range results {
		workoutIDSet[result.WorkoutID] = true
	}

	return workoutIDSet, nil
}

func (r *WorkoutRepo) GetLatestPersonalRecord(ctx context.Context, userID, exerciseID string) (*modelworkout.PersonalRecord, error) {
	var record personalRecordDocument
	err := r.personalRecordCollection.FindOne(
		ctx,
		bson.M{"user_id": userID, "exercise_id": exerciseID},
		options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: -1}}),
	).Decode(&record)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find personal record: %w", err)
	}

	return personalRecordModelFromDocument(&record), nil
}

func (r *WorkoutRepo) CreatePersonalRecord(ctx context.Context, record *modelworkout.PersonalRecord) error {
	if _, err := r.personalRecordCollection.InsertOne(ctx, personalRecordDocumentFromModel(record)); err != nil {
		return fmt.Errorf("insert personal record: %w", err)
	}

	return nil
}

func buildWorkoutListQuery(filter *modelworkout.ListFilter) bson.M {
	query := bson.M{}
	if len(filter.UserIDs) > 0 {
		query["user_id"] = bson.M{"$in": filter.UserIDs}
	} else {
		query["user_id"] = filter.UserID
	}
	if filter.Visibility != nil {
		query["visibility"] = *filter.Visibility
	}
	if filter.StartedAtGTE != nil || filter.StartedAtLT != nil {
		startedAtQuery := bson.M{}
		if filter.StartedAtGTE != nil {
			startedAtQuery["$gte"] = *filter.StartedAtGTE
		}
		if filter.StartedAtLT != nil {
			startedAtQuery["$lt"] = *filter.StartedAtLT
		}
		query["started_at"] = startedAtQuery
	}

	return query
}

func workoutDocumentFromModel(workout *modelworkout.Workout) *workoutDocument {
	return &workoutDocument{
		UserID:      workout.UserID,
		RoutineID:   workout.RoutineID,
		Title:       workout.Title,
		StartedAt:   workout.StartedAt,
		EndedAt:     workout.EndedAt,
		DurationSec: workout.DurationSec,
		Visibility:  workout.Visibility,
		Notes:       workout.Notes,
		Exercises:   workout.Exercises,
		Stats:       workout.Stats,
		CreatedAt:   workout.CreatedAt,
		UpdatedAt:   workout.UpdatedAt,
	}
}

func workoutModelFromDocument(doc *workoutDocument) *modelworkout.Workout {
	if doc == nil {
		return nil
	}

	return &modelworkout.Workout{
		ID:          doc.ID.Hex(),
		UserID:      doc.UserID,
		RoutineID:   doc.RoutineID,
		Title:       doc.Title,
		StartedAt:   doc.StartedAt,
		EndedAt:     doc.EndedAt,
		DurationSec: doc.DurationSec,
		Visibility:  doc.Visibility,
		Notes:       doc.Notes,
		Exercises:   doc.Exercises,
		Stats:       doc.Stats,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}
}

func personalRecordDocumentFromModel(record *modelworkout.PersonalRecord) *personalRecordDocument {
	return &personalRecordDocument{
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

func personalRecordModelFromDocument(doc *personalRecordDocument) *modelworkout.PersonalRecord {
	if doc == nil {
		return nil
	}

	return &modelworkout.PersonalRecord{
		ID:           doc.ID.Hex(),
		UserID:       doc.UserID,
		ExerciseID:   doc.ExerciseID,
		ExerciseName: doc.ExerciseName,
		BestWeightKG: doc.BestWeightKG,
		BestReps:     doc.BestReps,
		Estimated1RM: doc.Estimated1RM,
		WorkoutID:    doc.WorkoutID,
		UpdatedAt:    doc.UpdatedAt,
	}
}
