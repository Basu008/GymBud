package social

import "context"

type Repository interface {
	GetFollowCounts(ctx context.Context, userID string) (*FollowCounts, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	Follow(ctx context.Context, followerID, followeeID string) (string, error)
	AcceptFollowRequest(ctx context.Context, followerID, followeeID string) error
	RejectFollowRequest(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
}
