package social

import (
	"context"
	"errors"
	"strings"

	modelsocial "github.com/Basu008/GymBud/model/social"
	modeluser "github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/schema"
)

var ErrUserNotFound = errors.New("user not found")
var ErrCannotFollowYourself = errors.New("cannot follow yourself")
var ErrFollowRequestNotFound = errors.New("follow request not found")

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

func (s *Service) AcceptFollowRequest(ctx context.Context, currentUserID, requesterID string) (*schema.FollowActionResponse, error) {
	currentUserID = strings.TrimSpace(currentUserID)
	requesterID = strings.TrimSpace(requesterID)
	if currentUserID == requesterID {
		return nil, ErrCannotFollowYourself
	}

	requester, err := s.userRepo.GetByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := s.repo.AcceptFollowRequest(ctx, requesterID, currentUserID); err != nil {
		if errors.Is(err, modelsocial.ErrFollowRequestNotFound) {
			return nil, ErrFollowRequestNotFound
		}
		return nil, err
	}

	responseUser, err := s.buildUserResponse(ctx, requester)
	if err != nil {
		return nil, err
	}

	return &schema.FollowActionResponse{
		User:         responseUser,
		FollowStatus: modelsocial.FollowStatusAccepted,
	}, nil
}

func (s *Service) RejectFollowRequest(ctx context.Context, currentUserID, requesterID string) (*schema.FollowActionResponse, error) {
	currentUserID = strings.TrimSpace(currentUserID)
	requesterID = strings.TrimSpace(requesterID)
	if currentUserID == requesterID {
		return nil, ErrCannotFollowYourself
	}

	requester, err := s.userRepo.GetByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := s.repo.RejectFollowRequest(ctx, requesterID, currentUserID); err != nil {
		if errors.Is(err, modelsocial.ErrFollowRequestNotFound) {
			return nil, ErrFollowRequestNotFound
		}
		return nil, err
	}

	responseUser, err := s.buildUserResponse(ctx, requester)
	if err != nil {
		return nil, err
	}

	return &schema.FollowActionResponse{
		User:         responseUser,
		FollowStatus: modelsocial.FollowStatusNone,
	}, nil
}

func (s *Service) buildUserResponse(ctx context.Context, u *modeluser.User) (*schema.UserResponse, error) {
	counts, err := s.repo.GetFollowCounts(ctx, u.ID.String())
	if err != nil {
		return nil, err
	}
	bodyMetrics, err := s.userRepo.GetCurrentBodyMetrics(ctx, u.ID.String())
	if err != nil {
		if !errors.Is(err, modeluser.ErrBodyMetricsNotFound) {
			return nil, err
		}
		bodyMetrics = nil
	}

	return &schema.UserResponse{
		User:           u,
		BodyMetrics:    bodyMetrics,
		FollowersCount: counts.Followers,
		FollowingCount: counts.Following,
	}, nil
}
