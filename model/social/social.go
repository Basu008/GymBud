package social

import "errors"

const (
	FollowStatusAccepted = "accepted"
	FollowStatusPending  = "pending"
	FollowStatusNone     = "none"
)

var ErrFollowRequestNotFound = errors.New("follow request not found")

type FollowCounts struct {
	Followers int64
	Following int64
}

type WorkoutLikeSummary struct {
	WorkoutID  string
	LikesCount int
	LikedByMe  bool
}
