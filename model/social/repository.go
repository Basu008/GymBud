package social

import "context"

type Repository interface {
	GetFollowCounts(ctx context.Context, userID string) (*FollowCounts, error)
	Follow(ctx context.Context, followerID, followeeID string) (string, error)
	Unfollow(ctx context.Context, followerID, followeeID string) error
}
