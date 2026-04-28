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

	const createWorkoutLikesTable = `
		CREATE TABLE IF NOT EXISTS public.workout_likes (
			workout_id VARCHAR NOT NULL,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			PRIMARY KEY (workout_id, user_id)
		)
	`

	if _, err := r.db.Exec(ctx, createWorkoutLikesTable); err != nil {
		return fmt.Errorf("create workout_likes table: %w", err)
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

func (r *SocialRepo) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2 AND status = $3)`,
		followerID,
		followeeID,
		modelsocial.FollowStatusAccepted,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check following status: %w", err)
	}

	return exists, nil
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

func (r *SocialRepo) AcceptFollowRequest(ctx context.Context, followerID, followeeID string) error {
	result, err := r.db.Exec(
		ctx,
		`UPDATE follows SET status = $3 WHERE follower_id = $1 AND followee_id = $2 AND status = $4`,
		followerID,
		followeeID,
		modelsocial.FollowStatusAccepted,
		modelsocial.FollowStatusPending,
	)
	if err != nil {
		return fmt.Errorf("accept follow request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return modelsocial.ErrFollowRequestNotFound
	}

	return nil
}

func (r *SocialRepo) RejectFollowRequest(ctx context.Context, followerID, followeeID string) error {
	result, err := r.db.Exec(
		ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2 AND status = $3`,
		followerID,
		followeeID,
		modelsocial.FollowStatusPending,
	)
	if err != nil {
		return fmt.Errorf("reject follow request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return modelsocial.ErrFollowRequestNotFound
	}

	return nil
}

func (r *SocialRepo) LikeWorkout(ctx context.Context, workoutID, userID string) error {
	const query = `
		INSERT INTO workout_likes (workout_id, user_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (workout_id, user_id) DO NOTHING
	`

	if _, err := r.db.Exec(ctx, query, workoutID, userID); err != nil {
		return fmt.Errorf("like workout: %w", err)
	}

	return nil
}

func (r *SocialRepo) UnlikeWorkout(ctx context.Context, workoutID, userID string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM workout_likes WHERE workout_id = $1 AND user_id = $2`, workoutID, userID); err != nil {
		return fmt.Errorf("unlike workout: %w", err)
	}

	return nil
}

func (r *SocialRepo) DeleteWorkoutLikes(ctx context.Context, workoutID string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM workout_likes WHERE workout_id = $1`, workoutID); err != nil {
		return fmt.Errorf("delete workout likes: %w", err)
	}

	return nil
}

func (r *SocialRepo) GetWorkoutLikeSummaries(ctx context.Context, viewerUserID string, workoutIDs []string) (map[string]*modelsocial.WorkoutLikeSummary, error) {
	summaries := make(map[string]*modelsocial.WorkoutLikeSummary, len(workoutIDs))
	if len(workoutIDs) == 0 {
		return summaries, nil
	}

	const query = `
		SELECT
			workout_id,
			COUNT(*)::INT AS likes_count,
			BOOL_OR(user_id::TEXT = $2) AS liked_by_me
		FROM workout_likes
		WHERE workout_id = ANY($1)
		GROUP BY workout_id
	`

	rows, err := r.db.Query(ctx, query, workoutIDs, viewerUserID)
	if err != nil {
		return nil, fmt.Errorf("get workout like summaries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var summary modelsocial.WorkoutLikeSummary
		if err := rows.Scan(&summary.WorkoutID, &summary.LikesCount, &summary.LikedByMe); err != nil {
			return nil, fmt.Errorf("scan workout like summary: %w", err)
		}
		summaries[summary.WorkoutID] = &summary
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout like summaries: %w", err)
	}

	return summaries, nil
}
