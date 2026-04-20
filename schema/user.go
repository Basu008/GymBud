package schema

import (
	modeluser "github.com/Basu008/GymBud/model/user"
)

type SignUpUserBody struct {
	Username    string `json:"username" validate:"required"`
	Email       string `json:"email" validate:"required"`
	Password    string `json:"password" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
}

type LoginUserBody struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginUserResponse struct {
	User            *modeluser.User `json:"user"`
	RefreshToken    string          `json:"refresh_token"`
	AccessToken     string          `json:"access_token"`
	AccessTokenTTL  int64           `json:"access_token_ttl_seconds"`
	RefreshTokenTTL int64           `json:"refresh_token_ttl_seconds"`
}
