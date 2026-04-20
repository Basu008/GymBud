package user

import (
	"time"

	"github.com/google/uuid"
)

const UsersTable = "users"

type User struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Username        string     `db:"username" json:"username"`
	Email           string     `db:"email" json:"email"`
	PasswordHash    string     `db:"password_hash" json:"-"`
	DisplayName     string     `db:"display_name" json:"display_name,omitempty"`
	Plan            string     `db:"plan" json:"plan"`
	Bio             *string    `db:"bio" json:"bio,omitempty"`
	Gender          *string    `db:"gender" json:"gender,omitempty"`
	DateOfBirth     *time.Time `db:"date_of_birth" json:"date_of_birth,omitempty"`
	ProfileImageURL *string    `db:"profile_image_url" json:"profile_image_url,omitempty"`
	IsPrivate       bool       `db:"is_private" json:"is_private"`
	IsActive        bool       `db:"is_active" json:"is_active"`
	IsVerified      bool       `db:"is_verified" json:"is_verified"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}
