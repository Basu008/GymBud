package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Basu008/GymBud/model/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
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

	if _, err := r.db.Exec(ctx, `ALTER TABLE public.users ADD COLUMN IF NOT EXISTS plan VARCHAR NOT NULL DEFAULT 'free'`); err != nil {
		return fmt.Errorf("ensure users.plan column: %w", err)
	}

	return nil
}

func (r *UserRepo) Create(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is required")
	}

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

	var u user.User
	err := r.db.QueryRow(ctx, query, username).Scan(
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
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &u, nil
}
