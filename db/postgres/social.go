package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	modelsocial "github.com/Basu008/GymBud/model/social"
	modeluser "github.com/Basu008/GymBud/model/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SocialRepo struct {
	db *pgxpool.Pool
}

func NewSocialRepo(db *pgxpool.Pool) (*SocialRepo, error) {
	repo := &SocialRepo{db: db}
	if err := repo.initTable(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SocialRepo) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const createFollowsTable = `
		CREATE TABLE IF NOT EXISTS public.follows (
			follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			followee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status VARCHAR NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			PRIMARY KEY (follower_id, followee_id)
		)
	`

	if _, err := r.db.Exec(ctx, createFollowsTable); err != nil {
		return fmt.Errorf("create follows table: %w", err)
	}

	return nil
}

func (r *SocialRepo) GetFollowCounts(ctx context.Context, userID string) (*modelsocial.FollowCounts, error) {
	const query = `
		SELECT
			(SELECT COUNT(*) FROM follows WHERE followee_id = $1 AND status = $2) AS followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND status = $2) AS following_count
	`

	var counts modelsocial.FollowCounts
	if err := r.db.QueryRow(ctx, query, userID, modelsocial.FollowStatusAccepted).Scan(&counts.Followers, &counts.Following); err != nil {
		return nil, fmt.Errorf("get follow counts: %w", err)
	}

	return &counts, nil
}

func (r *SocialRepo) Follow(ctx context.Context, followerID, followeeID string) (string, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("begin follow transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var isPrivate bool
	if err := tx.QueryRow(ctx, `SELECT is_private FROM users WHERE id = $1`, followeeID).Scan(&isPrivate); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", modeluser.ErrUserNotFound
		}
		return "", fmt.Errorf("get followee privacy: %w", err)
	}

	status := modelsocial.FollowStatusAccepted
	if isPrivate {
		status = modelsocial.FollowStatusPending
	}

	const query = `
		INSERT INTO follows (
			follower_id,
			followee_id,
			status,
			created_at
		)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (follower_id, followee_id) DO UPDATE
		SET
			status = EXCLUDED.status,
			created_at = CASE
				WHEN follows.status = EXCLUDED.status THEN follows.created_at
				ELSE NOW()
			END
		RETURNING status
	`

	if err := tx.QueryRow(ctx, query, followerID, followeeID, status).Scan(&status); err != nil {
		return "", fmt.Errorf("upsert follow: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit follow transaction: %w", err)
	}

	return status, nil
}

func (r *SocialRepo) Unfollow(ctx context.Context, followerID, followeeID string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`, followerID, followeeID); err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}

	return nil
}
