package social

import "context"

type Repository interface {
	GetFollowCounts(ctx context.Context, userID string) (*FollowCounts, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	Follow(ctx context.Context, followerID, followeeID string) (string, error)
	AcceptFollowRequest(ctx context.Context, followerID, followeeID string) error
	RejectFollowRequest(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
	LikeWorkout(ctx context.Context, workoutID, userID string) error
	UnlikeWorkout(ctx context.Context, workoutID, userID string) error
	DeleteWorkoutLikes(ctx context.Context, workoutID string) error
	GetWorkoutLikeSummaries(ctx context.Context, viewerUserID string, workoutIDs []string) (map[string]*WorkoutLikeSummary, error)
}
