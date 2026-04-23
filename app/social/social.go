package social

import (
	"context"
	"errors"
	"strings"

	modeluser "github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/schema"
)

var ErrUserNotFound = errors.New("user not found")
var ErrCannotFollowYourself = errors.New("cannot follow yourself")

func (s *Service) FollowUser(ctx context.Context, followerID, followeeID string) (*schema.FollowActionResponse, error) {
	followerID = strings.TrimSpace(followerID)
	followeeID = strings.TrimSpace(followeeID)
	if followerID == followeeID {
		return nil, ErrCannotFollowYourself
	}

	targetUser, err := s.userRepo.GetByID(ctx, followeeID)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	status, err := s.repo.Follow(ctx, followerID, followeeID)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	responseUser, err := s.buildUserResponse(ctx, targetUser)
	if err != nil {
		return nil, err
	}

	return &schema.FollowActionResponse{
		User:         responseUser,
		FollowStatus: status,
	}, nil
}

func (s *Service) UnfollowUser(ctx context.Context, followerID, followeeID string) (*schema.UserResponse, error) {
	followerID = strings.TrimSpace(followerID)
	followeeID = strings.TrimSpace(followeeID)
	if followerID == followeeID {
		return nil, ErrCannotFollowYourself
	}

	targetUser, err := s.userRepo.GetByID(ctx, followeeID)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := s.repo.Unfollow(ctx, followerID, followeeID); err != nil {
		return nil, err
	}

	return s.buildUserResponse(ctx, targetUser)
}

func (s *Service) buildUserResponse(ctx context.Context, u *modeluser.User) (*schema.UserResponse, error) {
	counts, err := s.repo.GetFollowCounts(ctx, u.ID.String())
	if err != nil {
		return nil, err
	}

	return &schema.UserResponse{
		User:           u,
		FollowersCount: counts.Followers,
		FollowingCount: counts.Following,
	}, nil
}
