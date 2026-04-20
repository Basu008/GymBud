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

	return &schema.LoginUserResponse{
		User:           user,
		AccessToken:    session.AccessToken,
		AccessTokenTTL: session.AccessTokenTTL,
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (*modeluser.User, error) {
	foundUser, err := s.repo.GetByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return foundUser, nil
}

func (s *Service) UpdatePrivacy(ctx context.Context, userID string, isPrivate bool) (*modeluser.User, error) {
	foundUser, err := s.repo.UpdatePrivacyByID(ctx, strings.TrimSpace(userID), isPrivate)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return foundUser, nil
}

func (s *Service) UpdateActive(ctx context.Context, userID string, isActive bool) (*modeluser.User, error) {
	foundUser, err := s.repo.UpdateActiveByID(ctx, strings.TrimSpace(userID), isActive)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return foundUser, nil
}

func (s *Service) UpdateUser(ctx context.Context, userID string, body *schema.UpdateUserBody) (*modeluser.User, error) {
	updates := &modeluser.UserUpdate{}

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
	}

	user, err := s.repo.UpdateByID(ctx, strings.TrimSpace(userID), updates)
	if err != nil {
		if errors.Is(err, modeluser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
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
