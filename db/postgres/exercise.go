package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	modelexercise "github.com/Basu008/GymBud/model/exercise"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExerciseRepo struct {
	db *pgxpool.Pool
}

func NewExerciseRepo(db *pgxpool.Pool) (*ExerciseRepo, error) {
	repo := &ExerciseRepo{db: db}
	if err := repo.initTable(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *ExerciseRepo) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		CREATE TABLE IF NOT EXISTS public.exercises (
			id UUID PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			category VARCHAR(50) NOT NULL,
			is_made_by_admin BOOLEAN NOT NULL DEFAULT FALSE,
			equipment VARCHAR(50) NOT NULL,
			movement_mode VARCHAR(20),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT exercises_name_equipment_unique UNIQUE (name, equipment),
			CHECK (
				movement_mode IS NULL
				OR movement_mode IN ('unilateral', 'bilateral')
			)
		)
	`

	if _, err := r.db.Exec(ctx, query); err != nil {
		return fmt.Errorf("create exercises table: %w", err)
	}

	return nil
}

func (r *ExerciseRepo) Create(ctx context.Context, exercise *modelexercise.Exercise) error {
	const query = `
		INSERT INTO exercises (
			id,
			name,
			category,
			is_made_by_admin,
			equipment,
			movement_mode,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		exercise.ID,
		exercise.Name,
		exercise.Category,
		exercise.IsMadeByAdmin,
		exercise.Equipment,
		exercise.MovementMode,
		exercise.IsActive,
	).Scan(&exercise.CreatedAt, &exercise.UpdatedAt)
	if err != nil {
		if isExerciseNameConflict(err) {
			return modelexercise.ErrExerciseNameAlreadyExists
		}
		return fmt.Errorf("create exercise: %w", err)
	}

	return nil
}

func (r *ExerciseRepo) GetByID(ctx context.Context, exerciseID string) (*modelexercise.Exercise, error) {
	const query = `
		SELECT
			id,
			name,
			category,
			is_made_by_admin,
			equipment,
			movement_mode,
			is_active,
			created_at,
			updated_at
		FROM exercises
		WHERE id = $1
	`

	var exercise modelexercise.Exercise
	err := r.db.QueryRow(ctx, query, exerciseID).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Category,
		&exercise.IsMadeByAdmin,
		&exercise.Equipment,
		&exercise.MovementMode,
		&exercise.IsActive,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, modelexercise.ErrExerciseNotFound
		}
		return nil, fmt.Errorf("get exercise: %w", err)
	}

	return &exercise, nil
}

func (r *ExerciseRepo) UpdateByID(ctx context.Context, exerciseID string, input *modelexercise.UpdateInput) (*modelexercise.Exercise, error) {
	if input == nil {
		return nil, errors.New("exercise update is required")
	}

	var isMadeByAdmin bool
	if err := r.db.QueryRow(ctx, `SELECT is_made_by_admin FROM exercises WHERE id = $1`, exerciseID).Scan(&isMadeByAdmin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, modelexercise.ErrExerciseNotFound
		}
		return nil, fmt.Errorf("get exercise for update: %w", err)
	}
	if isMadeByAdmin {
		return nil, modelexercise.ErrExerciseManagedByAdmin
	}

	setClauses := make([]string, 0, 6)
	args := make([]any, 0, 7)
	argIndex := 1

	if input.NameSet {
		if input.Name == nil {
			return nil, errors.New("name cannot be null")
		}
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argIndex))
		args = append(args, *input.Name)
		argIndex++
	}
	if input.CategorySet {
		if input.Category == nil {
			return nil, errors.New("category cannot be null")
		}
		setClauses = append(setClauses, "category = $"+strconv.Itoa(argIndex))
		args = append(args, *input.Category)
		argIndex++
	}
	if input.EquipmentSet {
		if input.Equipment == nil {
			return nil, errors.New("equipment cannot be null")
		}
		setClauses = append(setClauses, "equipment = $"+strconv.Itoa(argIndex))
		args = append(args, *input.Equipment)
		argIndex++
	}
	if input.MovementModeSet {
		if input.MovementMode == nil {
			setClauses = append(setClauses, "movement_mode = NULL")
		} else {
			setClauses = append(setClauses, "movement_mode = $"+strconv.Itoa(argIndex))
			args = append(args, *input.MovementMode)
			argIndex++
		}
	}
	if input.IsMadeByAdminSet {
		if input.IsMadeByAdmin == nil {
			return nil, errors.New("is_admin cannot be null")
		}
		setClauses = append(setClauses, "is_made_by_admin = $"+strconv.Itoa(argIndex))
		args = append(args, *input.IsMadeByAdmin)
		argIndex++
	}
	if input.IsActiveSet {
		if input.IsActive == nil {
			return nil, errors.New("is_active cannot be null")
		}
		setClauses = append(setClauses, "is_active = $"+strconv.Itoa(argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil, errors.New("at least one updatable field is required")
	}

	query := `
		UPDATE exercises
		SET ` + strings.Join(append(setClauses, "updated_at = NOW()"), ", ") + `
		WHERE id = $` + strconv.Itoa(argIndex) + `
		RETURNING
			id,
			name,
			category,
			is_made_by_admin,
			equipment,
			movement_mode,
			is_active,
			created_at,
			updated_at
	`
	args = append(args, exerciseID)

	var exercise modelexercise.Exercise
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Category,
		&exercise.IsMadeByAdmin,
		&exercise.Equipment,
		&exercise.MovementMode,
		&exercise.IsActive,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, modelexercise.ErrExerciseNotFound
		}
		if isExerciseNameConflict(err) {
			return nil, modelexercise.ErrExerciseNameAlreadyExists
		}
		return nil, fmt.Errorf("update exercise: %w", err)
	}

	return &exercise, nil
}

func (r *ExerciseRepo) DeleteByID(ctx context.Context, exerciseID string) error {
	var isMadeByAdmin bool
	if err := r.db.QueryRow(ctx, `SELECT is_made_by_admin FROM exercises WHERE id = $1`, exerciseID).Scan(&isMadeByAdmin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return modelexercise.ErrExerciseNotFound
		}
		return fmt.Errorf("get exercise for delete: %w", err)
	}
	if isMadeByAdmin {
		return modelexercise.ErrExerciseManagedByAdmin
	}

	commandTag, err := r.db.Exec(ctx, `DELETE FROM exercises WHERE id = $1`, exerciseID)
	if err != nil {
		return fmt.Errorf("delete exercise: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return modelexercise.ErrExerciseNotFound
	}

	return nil
}

func isExerciseNameConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}

	constraintName := strings.ToLower(pgErr.ConstraintName)
	detail := strings.ToLower(pgErr.Detail)

	if constraintName == "exercises_name_equipment_unique" {
		return true
	}

	return strings.Contains(detail, "(name, equipment)") ||
		(strings.Contains(detail, "name") && strings.Contains(detail, "equipment"))
}
