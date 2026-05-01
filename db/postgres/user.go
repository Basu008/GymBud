package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Basu008/GymBud/model/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

type scanner interface {
	Scan(dest ...any) error
}

func NewUserRepo(db *pgxpool.Pool) (*UserRepo, error) {
	repo := &UserRepo{db: db}
	if err := repo.initTable(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *UserRepo) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := r.db.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`); err != nil {
		return fmt.Errorf("create pgcrypto extension: %w", err)
	}

	const query = `
		CREATE TABLE IF NOT EXISTS public.users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR UNIQUE NOT NULL,
			email VARCHAR UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			display_name VARCHAR NOT NULL,
			plan VARCHAR NOT NULL DEFAULT 'free',
			bio TEXT,
			gender VARCHAR(20),
			date_of_birth DATE,
			profile_image_url TEXT,
			is_private BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			is_verified BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	if _, err := r.db.Exec(ctx, query); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	const createBodyMetricsTable = `
		CREATE TABLE IF NOT EXISTS public.user_body_metrics (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			height_cm DECIMAL(5,2),
			weight_kg DECIMAL(5,2),
			recorded_at TIMESTAMP NOT NULL,
			source VARCHAR(30) DEFAULT 'manual',
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	if _, err := r.db.Exec(ctx, createBodyMetricsTable); err != nil {
		return fmt.Errorf("create user_body_metrics table: %w", err)
	}
	if _, err := r.db.Exec(ctx, `CREATE INDEX IF NOT EXISTS user_body_metrics_user_recorded_idx ON public.user_body_metrics (user_id, recorded_at DESC, created_at DESC, id DESC)`); err != nil {
		return fmt.Errorf("create user body metrics user recorded index: %w", err)
	}

	const createCurrentStatsTable = `
		CREATE TABLE IF NOT EXISTS public.user_current_stats (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			current_height_cm DECIMAL(5,2),
			current_weight_kg DECIMAL(5,2),
			bmi DECIMAL(5,2),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	if _, err := r.db.Exec(ctx, createCurrentStatsTable); err != nil {
		return fmt.Errorf("create user_current_stats table: %w", err)
	}

	return nil
}

func (r *UserRepo) Create(ctx context.Context, u *user.User) error {
	const query = `
		INSERT INTO users (
			username,
			email,
			password_hash,
			display_name,
			plan,
			bio,
			gender,
			date_of_birth,
			profile_image_url,
			is_private,
			is_active,
			is_verified
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at, plan, is_active, is_verified
	`

	err := r.db.QueryRow(
		ctx,
		query,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.DisplayName,
		u.Plan,
		u.Bio,
		u.Gender,
		u.DateOfBirth,
		u.ProfileImageURL,
		u.IsPrivate,
		u.IsActive,
		u.IsVerified,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt, &u.Plan, &u.IsActive, &u.IsVerified)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(strings.ToLower(pgErr.ConstraintName), "username") {
			return user.ErrUsernameAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	const query = `
		SELECT
			id,
			username,
			email,
			password_hash,
			display_name,
			plan,
			bio,
			gender,
			date_of_birth,
			profile_image_url,
			is_private,
			is_active,
			is_verified,
			created_at,
			updated_at
		FROM users
		WHERE username = $1
	`

	return r.getUserByQuery(ctx, query, username)
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*user.User, error) {
	const query = `
		SELECT
			id,
			username,
			email,
			password_hash,
			display_name,
			plan,
			bio,
			gender,
			date_of_birth,
			profile_image_url,
			is_private,
			is_active,
			is_verified,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	return r.getUserByQuery(ctx, query, id)
}

func (r *UserRepo) getUserByQuery(ctx context.Context, query string, arg any) (*user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, query, arg).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.Plan,
		&u.Bio,
		&u.Gender,
		&u.DateOfBirth,
		&u.ProfileImageURL,
		&u.IsPrivate,
		&u.IsActive,
		&u.IsVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) UpdateByID(ctx context.Context, userID string, updates *user.UserUpdate) (*user.User, error) {
	if updates == nil {
		return nil, errors.New("user update is required")
	}

	setClauses := make([]string, 0, 7)
	args := make([]any, 0, 8)
	argIndex := 1

	if updates.DisplayNameSet {
		if updates.DisplayName == nil {
			return nil, errors.New("display_name cannot be null")
		}
		setClauses = append(setClauses, "display_name = $"+strconv.Itoa(argIndex))
		args = append(args, *updates.DisplayName)
		argIndex++
	}
	if updates.PlanSet {
		if updates.Plan == nil {
			return nil, errors.New("plan cannot be null")
		}
		setClauses = append(setClauses, "plan = $"+strconv.Itoa(argIndex))
		args = append(args, *updates.Plan)
		argIndex++
	}
	if updates.BioSet {
		if updates.Bio == nil {
			setClauses = append(setClauses, "bio = NULL")
		} else {
			setClauses = append(setClauses, "bio = $"+strconv.Itoa(argIndex))
			args = append(args, *updates.Bio)
			argIndex++
		}
	}
	if updates.GenderSet {
		if updates.Gender == nil {
			setClauses = append(setClauses, "gender = NULL")
		} else {
			setClauses = append(setClauses, "gender = $"+strconv.Itoa(argIndex))
			args = append(args, *updates.Gender)
			argIndex++
		}
	}
	if updates.DateOfBirthSet {
		if updates.DateOfBirth == nil {
			setClauses = append(setClauses, "date_of_birth = NULL")
		} else {
			setClauses = append(setClauses, "date_of_birth = $"+strconv.Itoa(argIndex))
			args = append(args, *updates.DateOfBirth)
			argIndex++
		}
	}
	if updates.ProfileImageSet {
		if updates.ProfileImageURL == nil {
			setClauses = append(setClauses, "profile_image_url = NULL")
		} else {
			setClauses = append(setClauses, "profile_image_url = $"+strconv.Itoa(argIndex))
			args = append(args, *updates.ProfileImageURL)
			argIndex++
		}
	}

	if len(setClauses) == 0 {
		return nil, errors.New("at least one updatable field is required")
	}

	args = append(args, userID)
	query := `
		UPDATE users
		SET ` + strings.Join(append(setClauses, "updated_at = NOW()"), ", ") + `
		WHERE id = $` + strconv.Itoa(argIndex) + `
		RETURNING
			id,
			username,
			email,
			password_hash,
			display_name,
			plan,
			bio,
			gender,
			date_of_birth,
			profile_image_url,
			is_private,
			is_active,
			is_verified,
			created_at,
			updated_at
	`

	return r.getUpdatedUser(ctx, query, args...)
}

func (r *UserRepo) UpdatePrivacyByID(ctx context.Context, userID string, isPrivate bool) (*user.User, error) {
	return r.updateUserFlagByID(ctx, userID, "is_private", isPrivate)
}

func (r *UserRepo) UpdateActiveByID(ctx context.Context, userID string, isActive bool) (*user.User, error) {
	return r.updateUserFlagByID(ctx, userID, "is_active", isActive)
}

func (r *UserRepo) updateUserFlagByID(ctx context.Context, userID, column string, value bool) (*user.User, error) {
	query := `
		UPDATE users
		SET ` + column + ` = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING
			id,
			username,
			email,
			password_hash,
			display_name,
			plan,
			bio,
			gender,
			date_of_birth,
			profile_image_url,
			is_private,
			is_active,
			is_verified,
			created_at,
			updated_at
	`

	return r.getUpdatedUser(ctx, query, value, userID)
}

func (r *UserRepo) getUpdatedUser(ctx context.Context, query string, args ...any) (*user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.Plan,
		&u.Bio,
		&u.Gender,
		&u.DateOfBirth,
		&u.ProfileImageURL,
		&u.IsPrivate,
		&u.IsActive,
		&u.IsVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) CreateBodyMetrics(ctx context.Context, metrics *user.BodyMetrics) (*user.CurrentStats, error) {
	if metrics == nil {
		return nil, errors.New("body metrics is required")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin create body metrics transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const insertQuery = `
		INSERT INTO user_body_metrics (
			id,
			user_id,
			height_cm,
			weight_kg,
			recorded_at,
			source
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	if err := tx.QueryRow(
		ctx,
		insertQuery,
		metrics.ID,
		metrics.UserID,
		metrics.HeightCM,
		metrics.WeightKG,
		metrics.RecordedAt,
		metrics.Source,
	).Scan(&metrics.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert user body metrics: %w", err)
	}

	currentStats, err := r.refreshCurrentStats(ctx, tx, metrics.UserID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create body metrics transaction: %w", err)
	}

	return currentStats, nil
}

func (r *UserRepo) GetCurrentBodyMetrics(ctx context.Context, userID string) (*user.BodyMetrics, error) {
	const query = `
		SELECT
			id,
			user_id,
			height_cm,
			weight_kg,
			recorded_at,
			source,
			created_at
		FROM user_body_metrics
		WHERE user_id = $1
		ORDER BY recorded_at DESC, created_at DESC, id DESC
		LIMIT 1
	`

	metrics, err := r.getBodyMetricsByQuery(ctx, query, userID)
	if err != nil {
		if errors.Is(err, user.ErrBodyMetricsNotFound) {
			return nil, user.ErrBodyMetricsNotFound
		}
		return nil, fmt.Errorf("get current body metrics: %w", err)
	}

	return metrics, nil
}

func (r *UserRepo) ListBodyMetrics(ctx context.Context, userID string, limit int) ([]*user.BodyMetrics, error) {
	if limit <= 0 {
		return []*user.BodyMetrics{}, nil
	}

	const query = `
		SELECT
			id,
			user_id,
			height_cm,
			weight_kg,
			recorded_at,
			source,
			created_at
		FROM user_body_metrics
		WHERE user_id = $1
		ORDER BY recorded_at DESC, created_at DESC, id DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list body metrics: %w", err)
	}
	defer rows.Close()

	metrics := make([]*user.BodyMetrics, 0, limit)
	for rows.Next() {
		var m user.BodyMetrics
		if err := scanBodyMetrics(rows, &m); err != nil {
			return nil, fmt.Errorf("scan body metrics: %w", err)
		}
		metrics = append(metrics, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate body metrics: %w", err)
	}

	return metrics, nil
}

func (r *UserRepo) getBodyMetricsByQuery(ctx context.Context, query string, args ...any) (*user.BodyMetrics, error) {
	var metrics user.BodyMetrics
	if err := scanBodyMetrics(r.db.QueryRow(ctx, query, args...), &metrics); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrBodyMetricsNotFound
		}
		return nil, err
	}

	return &metrics, nil
}

func scanBodyMetrics(row scanner, metrics *user.BodyMetrics) error {
	return row.Scan(
		&metrics.ID,
		&metrics.UserID,
		&metrics.HeightCM,
		&metrics.WeightKG,
		&metrics.RecordedAt,
		&metrics.Source,
		&metrics.CreatedAt,
	)
}

func (r *UserRepo) DeleteBodyMetrics(ctx context.Context, userID, metricsID string) (*user.CurrentStats, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin delete body metrics transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	commandTag, err := tx.Exec(ctx, `DELETE FROM user_body_metrics WHERE id = $1 AND user_id = $2`, metricsID, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("delete user body metrics: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return nil, user.ErrBodyMetricsNotFound
	}

	currentStats, err := r.refreshCurrentStats(ctx, tx, parsedUserID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit delete body metrics transaction: %w", err)
	}

	return currentStats, nil
}

func (r *UserRepo) refreshCurrentStats(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*user.CurrentStats, error) {
	const latestMetricsQuery = `
		SELECT
			height_cm,
			weight_kg
		FROM user_body_metrics
		WHERE user_id = $1
		ORDER BY recorded_at DESC, created_at DESC, id DESC
		LIMIT 1
	`

	var stats user.CurrentStats
	stats.UserID = userID

	err := tx.QueryRow(ctx, latestMetricsQuery, userID).Scan(
		&stats.CurrentHeightCM,
		&stats.CurrentWeightKG,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if _, deleteErr := tx.Exec(ctx, `DELETE FROM user_current_stats WHERE user_id = $1`, userID); deleteErr != nil {
				return nil, fmt.Errorf("delete user current stats: %w", deleteErr)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("get latest body metrics: %w", err)
	}

	stats.BMI = calculateBMI(stats.CurrentHeightCM, stats.CurrentWeightKG)

	const upsertQuery = `
		INSERT INTO user_current_stats (
			user_id,
			current_height_cm,
			current_weight_kg,
			bmi,
			updated_at
		)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET
			current_height_cm = EXCLUDED.current_height_cm,
			current_weight_kg = EXCLUDED.current_weight_kg,
			bmi = EXCLUDED.bmi,
			updated_at = NOW()
		RETURNING updated_at
	`

	if err := tx.QueryRow(
		ctx,
		upsertQuery,
		stats.UserID,
		stats.CurrentHeightCM,
		stats.CurrentWeightKG,
		stats.BMI,
	).Scan(&stats.UpdatedAt); err != nil {
		return nil, fmt.Errorf("upsert user current stats: %w", err)
	}

	return &stats, nil
}

func calculateBMI(heightCM, weightKG *float64) *float64 {
	if heightCM == nil || weightKG == nil || *heightCM <= 0 || *weightKG <= 0 {
		return nil
	}

	heightM := *heightCM / 100
	bmi := *weightKG / (heightM * heightM)
	bmi = float64(int(bmi*100+0.5)) / 100
	return &bmi
}
