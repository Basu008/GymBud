package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	modelexercise "github.com/Basu008/GymBud/model/exercise"
	modelroutine "github.com/Basu008/GymBud/model/routine"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoutineRepo struct {
	db *pgxpool.Pool
}

func NewRoutineRepo(db *pgxpool.Pool) (*RoutineRepo, error) {
	repo := &RoutineRepo{db: db}
	if err := repo.initTable(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *RoutineRepo) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const createRoutinesTable = `
		CREATE TABLE IF NOT EXISTS public.routines (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`
	if _, err := r.db.Exec(ctx, createRoutinesTable); err != nil {
		return fmt.Errorf("create routines table: %w", err)
	}

	const createRoutineExercisesTable = `
		CREATE TABLE IF NOT EXISTS public.routine_exercises (
			id UUID PRIMARY KEY,
			routine_id UUID NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
			exercise_id UUID NOT NULL REFERENCES exercises(id),
			order_index INT NOT NULL,
			notes TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CHECK (order_index > 0),
			UNIQUE (routine_id, order_index)
		)
	`
	if _, err := r.db.Exec(ctx, createRoutineExercisesTable); err != nil {
		return fmt.Errorf("create routine_exercises table: %w", err)
	}

	const createRoutineExerciseSetsTable = `
		CREATE TABLE IF NOT EXISTS public.routine_exercise_sets (
			id UUID PRIMARY KEY,
			routine_exercise_id UUID NOT NULL REFERENCES routine_exercises(id) ON DELETE CASCADE,
			set_number INT NOT NULL,
			min_reps INT NOT NULL,
			max_reps INT NOT NULL,
			target_weight_kg DECIMAL(6,2),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CHECK (set_number > 0),
			CHECK (min_reps > 0),
			CHECK (max_reps >= min_reps),
			CHECK (target_weight_kg IS NULL OR target_weight_kg >= 0),
			UNIQUE (routine_exercise_id, set_number)
		)
	`
	if _, err := r.db.Exec(ctx, createRoutineExerciseSetsTable); err != nil {
		return fmt.Errorf("create routine_exercise_sets table: %w", err)
	}

	return nil
}

func (r *RoutineRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM routines WHERE user_id = $1`, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count routines: %w", err)
	}
	return count, nil
}

func (r *RoutineRepo) Create(ctx context.Context, routine *modelroutine.Routine) (*modelroutine.Routine, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin create routine transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(
		ctx,
		`INSERT INTO routines (id, user_id, name) VALUES ($1, $2, $3) RETURNING created_at, updated_at`,
		routine.ID,
		routine.UserID,
		routine.Name,
	).Scan(&routine.CreatedAt, &routine.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create routine: %w", err)
	}

	if err := r.insertRoutineExercises(ctx, tx, routine); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create routine transaction: %w", err)
	}

	return r.GetByID(ctx, routine.UserID, routine.ID)
}

func (r *RoutineRepo) ListByUserID(ctx context.Context, userID string) ([]*modelroutine.Routine, error) {
	rows, err := r.db.Query(ctx, `SELECT id FROM routines WHERE user_id = $1 ORDER BY created_at ASC, name ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list routine ids: %w", err)
	}
	defer rows.Close()

	routines := make([]*modelroutine.Routine, 0)
	for rows.Next() {
		var routineID string
		if err := rows.Scan(&routineID); err != nil {
			return nil, fmt.Errorf("scan routine id: %w", err)
		}

		routine, err := r.GetByID(ctx, userID, routineID)
		if err != nil {
			return nil, err
		}
		routines = append(routines, routine)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate routine ids: %w", err)
	}

	return routines, nil
}

func (r *RoutineRepo) GetByID(ctx context.Context, userID, routineID string) (*modelroutine.Routine, error) {
	return r.getByConditions(ctx, `WHERE id = $1 AND user_id = $2`, routineID, userID)
}

func (r *RoutineRepo) GetByRoutineID(ctx context.Context, routineID string) (*modelroutine.Routine, error) {
	return r.getByConditions(ctx, `WHERE id = $1`, routineID)
}

func (r *RoutineRepo) getByConditions(ctx context.Context, whereClause string, args ...any) (*modelroutine.Routine, error) {
	var routine modelroutine.Routine
	err := r.db.QueryRow(
		ctx,
		`SELECT id, user_id, name, created_at, updated_at FROM routines `+whereClause,
		args...,
	).Scan(&routine.ID, &routine.UserID, &routine.Name, &routine.CreatedAt, &routine.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, modelroutine.ErrRoutineNotFound
		}
		return nil, fmt.Errorf("get routine: %w", err)
	}

	exerciseRows, err := r.db.Query(ctx, `
		SELECT
			re.id,
			re.routine_id,
			re.exercise_id,
			re.order_index,
			re.notes,
			re.created_at,
			re.updated_at,
			e.id,
			e.name,
			e.slug,
			e.category,
			e.user_id,
			e.is_made_by_admin,
			e.equipment,
			e.primary_muscle,
			e.secondary_muscles,
			e.difficulty,
			e.movement_mode,
			e.is_active,
			e.created_at,
			e.updated_at
		FROM routine_exercises re
		JOIN exercises e ON e.id = re.exercise_id
		WHERE re.routine_id = $1
		ORDER BY re.order_index ASC, re.created_at ASC
	`, routine.ID)
	if err != nil {
		return nil, fmt.Errorf("query routine exercises: %w", err)
	}
	defer exerciseRows.Close()

	routineExercises := make([]*modelroutine.RoutineExercise, 0)
	routineExerciseByID := make(map[string]*modelroutine.RoutineExercise)
	routineExerciseIDs := make([]string, 0)

	for exerciseRows.Next() {
		var routineExercise modelroutine.RoutineExercise
		var exercise modelexercise.Exercise
		if err := exerciseRows.Scan(
			&routineExercise.ID,
			&routineExercise.RoutineID,
			&routineExercise.ExerciseID,
			&routineExercise.OrderIndex,
			&routineExercise.Notes,
			&routineExercise.CreatedAt,
			&routineExercise.UpdatedAt,
			&exercise.ID,
			&exercise.Name,
			&exercise.Slug,
			&exercise.Category,
			&exercise.UserID,
			&exercise.IsMadeByAdmin,
			&exercise.Equipment,
			&exercise.PrimaryMuscle,
			&exercise.SecondaryMuscles,
			&exercise.Difficulty,
			&exercise.MovementMode,
			&exercise.IsActive,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan routine exercise: %w", err)
		}

		routineExercise.Exercise = &exercise
		routineExercise.Sets = make([]*modelroutine.RoutineExerciseSet, 0)

		routineExerciseCopy := routineExercise
		routineExercises = append(routineExercises, &routineExerciseCopy)
		routineExerciseByID[routineExerciseCopy.ID] = &routineExerciseCopy
		routineExerciseIDs = append(routineExerciseIDs, routineExerciseCopy.ID)
	}
	if err := exerciseRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate routine exercises: %w", err)
	}

	if len(routineExerciseIDs) > 0 {
		setRows, err := r.db.Query(ctx, `
			SELECT
				id,
				routine_exercise_id,
				set_number,
				min_reps,
				max_reps,
				target_weight_kg,
				created_at,
				updated_at
			FROM routine_exercise_sets
			WHERE routine_exercise_id = ANY($1)
			ORDER BY routine_exercise_id ASC, set_number ASC
		`, routineExerciseIDs)
		if err != nil {
			return nil, fmt.Errorf("query routine exercise sets: %w", err)
		}
		defer setRows.Close()

		for setRows.Next() {
			var set modelroutine.RoutineExerciseSet
			if err := setRows.Scan(
				&set.ID,
				&set.RoutineExerciseID,
				&set.SetNumber,
				&set.MinReps,
				&set.MaxReps,
				&set.TargetWeightKG,
				&set.CreatedAt,
				&set.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("scan routine exercise set: %w", err)
			}

			routineExercise, ok := routineExerciseByID[set.RoutineExerciseID]
			if ok {
				routineExercise.Sets = append(routineExercise.Sets, &set)
			}
		}
		if err := setRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate routine exercise sets: %w", err)
		}
	}

	routine.Exercises = routineExercises
	return &routine, nil
}

func (r *RoutineRepo) ReplaceByID(ctx context.Context, userID string, routine *modelroutine.Routine) (*modelroutine.Routine, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin update routine transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	commandTag, err := tx.Exec(ctx, `UPDATE routines SET name = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`, routine.Name, routine.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("update routine: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return nil, modelroutine.ErrRoutineNotFound
	}

	if _, err := tx.Exec(ctx, `DELETE FROM routine_exercises WHERE routine_id = $1`, routine.ID); err != nil {
		return nil, fmt.Errorf("delete routine exercises: %w", err)
	}

	if err := r.insertRoutineExercises(ctx, tx, routine); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update routine transaction: %w", err)
	}

	return r.GetByID(ctx, userID, routine.ID)
}

func (r *RoutineRepo) DeleteByID(ctx context.Context, userID, routineID string) error {
	commandTag, err := r.db.Exec(ctx, `DELETE FROM routines WHERE id = $1 AND user_id = $2`, routineID, userID)
	if err != nil {
		return fmt.Errorf("delete routine: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return modelroutine.ErrRoutineNotFound
	}

	return nil
}

func (r *RoutineRepo) insertRoutineExercises(ctx context.Context, tx pgx.Tx, routine *modelroutine.Routine) error {
	for _, exercise := range routine.Exercises {
		exercise.RoutineID = routine.ID
		if err := tx.QueryRow(
			ctx,
			`INSERT INTO routine_exercises (id, routine_id, exercise_id, order_index, notes) VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`,
			exercise.ID,
			exercise.RoutineID,
			exercise.ExerciseID,
			exercise.OrderIndex,
			exercise.Notes,
		).Scan(&exercise.CreatedAt, &exercise.UpdatedAt); err != nil {
			if isRoutineExerciseForeignKeyError(err) {
				return modelroutine.ErrRoutineExerciseNotFound
			}
			return fmt.Errorf("insert routine exercise: %w", err)
		}

		for _, set := range exercise.Sets {
			set.RoutineExerciseID = exercise.ID
			if err := tx.QueryRow(
				ctx,
				`INSERT INTO routine_exercise_sets (id, routine_exercise_id, set_number, min_reps, max_reps, target_weight_kg) VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`,
				set.ID,
				set.RoutineExerciseID,
				set.SetNumber,
				set.MinReps,
				set.MaxReps,
				set.TargetWeightKG,
			).Scan(&set.CreatedAt, &set.UpdatedAt); err != nil {
				return fmt.Errorf("insert routine exercise set: %w", err)
			}
		}
	}

	return nil
}

func isRoutineExerciseForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return false
	}

	constraintName := strings.ToLower(pgErr.ConstraintName)
	return strings.Contains(constraintName, "exercise_id") || strings.Contains(constraintName, "routine_exercises_exercise_id_fkey")
}
