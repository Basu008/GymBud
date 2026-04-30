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
			slug VARCHAR(120) NOT NULL,
			category VARCHAR(50) NOT NULL,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			is_made_by_admin BOOLEAN NOT NULL DEFAULT FALSE,
			equipment VARCHAR(50) NOT NULL,
			primary_muscle VARCHAR(50) NOT NULL,
			secondary_muscles TEXT[] NOT NULL DEFAULT '{}',
			difficulty VARCHAR(30) NOT NULL,
			movement_mode VARCHAR(20),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CHECK (
				movement_mode IS NULL
				OR movement_mode IN ('unilateral', 'bilateral')
			),
			CHECK (is_made_by_admin OR user_id IS NOT NULL)
		)
	`

	if _, err := r.db.Exec(ctx, query); err != nil {
		return fmt.Errorf("create exercises table: %w", err)
	}
	if _, err := r.db.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS unique_admin_exercise ON public.exercises (slug, equipment) WHERE is_made_by_admin = TRUE`); err != nil {
		return fmt.Errorf("create unique admin exercise index: %w", err)
	}
	if _, err := r.db.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS unique_user_exercise ON public.exercises (user_id, slug, equipment) WHERE is_made_by_admin = FALSE`); err != nil {
		return fmt.Errorf("create unique user exercise index: %w", err)
	}

	return nil
}

func (r *ExerciseRepo) Create(ctx context.Context, exercise *modelexercise.Exercise) error {
	const query = `
		INSERT INTO exercises (
			id,
			name,
			slug,
			category,
			user_id,
			is_made_by_admin,
			equipment,
			primary_muscle,
			secondary_muscles,
			difficulty,
			movement_mode,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		exercise.ID,
		exercise.Name,
		exercise.Slug,
		exercise.Category,
		exercise.UserID,
		exercise.IsMadeByAdmin,
		exercise.Equipment,
		exercise.PrimaryMuscle,
		exercise.SecondaryMuscles,
		exercise.Difficulty,
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

func (r *ExerciseRepo) List(ctx context.Context, filter *modelexercise.ListFilter) ([]*modelexercise.Exercise, error) {
	query := `
		SELECT
			id,
			name,
			slug,
			category,
			user_id,
			is_made_by_admin,
			equipment,
			primary_muscle,
			secondary_muscles,
			difficulty,
			movement_mode,
			is_active,
			created_at,
			updated_at
		FROM exercises
	`

	whereClauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	argIndex := 1

	if filter != nil {
		if strings.TrimSpace(filter.UserID) != "" {
			whereClauses = append(whereClauses, "(is_made_by_admin = TRUE OR user_id = $"+strconv.Itoa(argIndex)+")")
			args = append(args, strings.TrimSpace(filter.UserID))
			argIndex++
		}
		if filter.NameRegex != nil {
			whereClauses = append(whereClauses, "name ~* $"+strconv.Itoa(argIndex))
			args = append(args, *filter.NameRegex)
			argIndex++
		}
		if filter.Category != nil {
			whereClauses = append(whereClauses, "LOWER(category) = $"+strconv.Itoa(argIndex))
			args = append(args, *filter.Category)
			argIndex++
		}
	}

	if len(whereClauses) > 0 {
		query += "\nWHERE " + strings.Join(whereClauses, " AND ")
	}

	query += "\nORDER BY name ASC, created_at ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	exercises := make([]*modelexercise.Exercise, 0)
	for rows.Next() {
		var exercise modelexercise.Exercise
		if err := rows.Scan(
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
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		exercises = append(exercises, &exercise)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercises: %w", err)
	}

	return exercises, nil
}

func (r *ExerciseRepo) GetByID(ctx context.Context, exerciseID string) (*modelexercise.Exercise, error) {
	const query = `
		SELECT
			id,
			name,
			slug,
			category,
			user_id,
			is_made_by_admin,
			equipment,
			primary_muscle,
			secondary_muscles,
			difficulty,
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

	setClauses := make([]string, 0, 10)
	args := make([]any, 0, 11)
	argIndex := 1

	if input.NameSet {
		if input.Name == nil {
			return nil, errors.New("name cannot be null")
		}
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argIndex))
		args = append(args, *input.Name)
		argIndex++
	}
	if input.SlugSet {
		if input.Slug == nil {
			return nil, errors.New("slug cannot be null")
		}
		setClauses = append(setClauses, "slug = $"+strconv.Itoa(argIndex))
		args = append(args, *input.Slug)
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
	if input.PrimaryMuscleSet {
		if input.PrimaryMuscle == nil {
			return nil, errors.New("primary_muscle cannot be null")
		}
		setClauses = append(setClauses, "primary_muscle = $"+strconv.Itoa(argIndex))
		args = append(args, *input.PrimaryMuscle)
		argIndex++
	}
	if input.SecondaryMusclesSet {
		setClauses = append(setClauses, "secondary_muscles = $"+strconv.Itoa(argIndex))
		args = append(args, input.SecondaryMuscles)
		argIndex++
	}
	if input.DifficultySet {
		if input.Difficulty == nil {
			return nil, errors.New("difficulty cannot be null")
		}
		setClauses = append(setClauses, "difficulty = $"+strconv.Itoa(argIndex))
		args = append(args, *input.Difficulty)
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
			slug,
			category,
			user_id,
			is_made_by_admin,
			equipment,
			primary_muscle,
			secondary_muscles,
			difficulty,
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

	if constraintName == "unique_admin_exercise" || constraintName == "unique_user_exercise" {
		return true
	}

	return strings.Contains(detail, "(slug, equipment)") ||
		strings.Contains(detail, "(user_id, slug, equipment)") ||
		(strings.Contains(detail, "slug") && strings.Contains(detail, "equipment"))
}
