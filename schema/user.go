package schema

import (
	"time"

	modeluser "github.com/Basu008/GymBud/model/user"
)

type SignUpUserBody struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginUserBody struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateUserBody struct {
	DisplayName     *string `json:"display_name"`
	Plan            *string `json:"plan"`
	Bio             *string `json:"bio"`
	Gender          *string `json:"gender" validate:"oneof=M F"`
	DateOfBirth     *string `json:"date_of_birth"`
	ProfileImageURL *string `json:"profile_image_url"`
}

type UpdateUserInput struct {
	DisplayName     *string
	Plan            *string
	Bio             *string
	Gender          *string
	DateOfBirth     *time.Time
	ProfileImageURL *string
}

type UpdatePrivacyBody struct {
	IsPrivate bool `json:"is_private"`
}

type UpdateActiveBody struct {
	IsActive bool `json:"is_active"`
}

type UserResponse struct {
	*modeluser.User
	BodyMetrics    *modeluser.BodyMetrics `json:"body_metrics,omitempty"`
	FollowersCount int64                  `json:"followers_count"`
	FollowingCount int64                  `json:"following_count"`
}

type FollowActionResponse struct {
	User         *UserResponse `json:"user"`
	FollowStatus string        `json:"follow_status,omitempty"`
}

type LoginUserResponse struct {
	User        *UserResponse `json:"user"`
	AccessToken string        `json:"access_token"`
}

type DeleteAccountBody struct {
	Reason string `json:"reason" validate:"required"`
}
