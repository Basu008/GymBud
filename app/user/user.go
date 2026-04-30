package user

import (
	"context"
	"errors"
	"strings"
	"time"

	modeluser "github.com/Basu008/GymBud/model/user"
	"github.com/Basu008/GymBud/schema"
	"golang.org/x/crypto/bcrypt"
)

var ErrUsernameAlreadyExists = errors.New("username already exists")
var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrUserNotFound = errors.New("user not found")
var ErrUserInactive = errors.New("user is inactive")

func (s *Service) SignUpUser(ctx context.Context, opts *schema.SignUpUserBody) error {
	username := strings.TrimSpace(opts.Username)
	email := strings.TrimSpace(opts.Email)
	password := strings.TrimSpace(opts.Password)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := modeluser.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		DisplayName:  "",
		Plan:         "free",
		IsActive:     true,
		IsVerified:   false,
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		if errors.Is(err, modeluser.ErrUsernameAlreadyExists) {
			return ErrUsernameAlreadyExists
		}
		return err
	}

	return nil
}

func (s *Service) LoginUser(ctx context.Context, opts *schema.LoginUserBody) (*schema.LoginUserResponse, error) {
	username := strings.TrimSpace(opts.Username)
	password := strings.TrimSpace(opts.Password)

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	session, err := s.authService.CreateLoginSession(ctx, user.ID.String(), user.Plan)
	if err != nil {
		return nil, err
	}

	responseUser, err := s.buildUserResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	return &schema.LoginUserResponse{
		User:        responseUser,
		AccessToken: session.AccessToken,
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (*schema.UserResponse, error) {
	foundUser, err := s.repo.GetByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return s.buildUserResponse(ctx, foundUser)
}

func (s *Service) UpdatePrivacy(ctx context.Context, userID string, isPrivate bool) (*schema.UserResponse, error) {
	foundUser, err := s.repo.UpdatePrivacyByID(ctx, strings.TrimSpace(userID), isPrivate)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return s.buildUserResponse(ctx, foundUser)
}

func (s *Service) UpdateActive(ctx context.Context, userID string, isActive bool) (*schema.UserResponse, error) {
	foundUser, err := s.repo.UpdateActiveByID(ctx, strings.TrimSpace(userID), isActive)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return s.buildUserResponse(ctx, foundUser)
}

func (s *Service) UpdateUser(ctx context.Context, userID string, body *schema.UpdateUserBody) (*schema.UserResponse, error) {
	updates := &modeluser.UserUpdate{}
	userID = strings.TrimSpace(userID)
	var previousProfileImageURL *string

	if body.DisplayName != nil {
		value := strings.TrimSpace(*body.DisplayName)
		updates.DisplayNameSet = true
		updates.DisplayName = &value
	}
	if body.Plan != nil {
		value := strings.TrimSpace(*body.Plan)
		if value == "" {
			return nil, errors.New("plan cannot be empty")
		}
		updates.PlanSet = true
		updates.Plan = &value
	}
	if body.Bio != nil {
		value := strings.TrimSpace(*body.Bio)
		updates.BioSet = true
		updates.Bio = normalizeOptionalString(value)
	}
	if body.Gender != nil {
		value := strings.TrimSpace(*body.Gender)
		updates.GenderSet = true
		updates.Gender = normalizeOptionalString(value)
	}
	if body.DateOfBirth != nil {
		updates.DateOfBirthSet = true
		value := strings.TrimSpace(*body.DateOfBirth)
		if value != "" {
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil {
				return nil, errors.New("date_of_birth must be in YYYY-MM-DD format")
			}
			updates.DateOfBirth = &parsed
		}
	}
	if body.ProfileImageURL != nil {
		value := strings.TrimSpace(*body.ProfileImageURL)
		updates.ProfileImageSet = true
		updates.ProfileImageURL = normalizeOptionalString(value)

		currentUser, err := s.repo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, modeluser.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		previousProfileImageURL = currentUser.ProfileImageURL
	}

	user, err := s.repo.UpdateByID(ctx, userID, updates)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if shouldDeletePreviousProfileImage(previousProfileImageURL, updates.ProfileImageURL) {
		if err := s.mediaRepo.DeleteByImageURL(ctx, *previousProfileImageURL); err != nil {
			return nil, err
		}
	}

	return s.buildUserResponse(ctx, user)
}

func shouldDeletePreviousProfileImage(previous, next *string) bool {
	if previous == nil || strings.TrimSpace(*previous) == "" {
		return false
	}
	if next == nil {
		return true
	}
	return strings.TrimSpace(*previous) != strings.TrimSpace(*next)
}

func normalizeOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.authService.LogoutSession(ctx, strings.TrimSpace(sessionID))
}

func (s *Service) buildUserResponse(ctx context.Context, u *modeluser.User) (*schema.UserResponse, error) {
	counts, err := s.socialRepo.GetFollowCounts(ctx, u.ID.String())
	if err != nil {
		return nil, err
	}

	return &schema.UserResponse{
		User:           u,
		FollowersCount: counts.Followers,
		FollowingCount: counts.Following,
	}, nil
}
