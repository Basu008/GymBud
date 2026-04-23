package social

const (
	FollowStatusAccepted = "accepted"
	FollowStatusPending  = "pending"
)

type FollowCounts struct {
	Followers int64
	Following int64
}
