package user

import (
	"time"

	"github.com/google/uuid"
)

const UsersTable = "users"

type User struct {
	ID              uuid.UUID `db:"id" json:"id"`
	Username        string    `db:"username" json:"username"`
	Email           string    `db:"email" json:"email"`
	PasswordHash    string    `db:"password_hash" json:"-"`
	DisplayName     *string   `db:"display_name" json:"display_name,omitempty"`
	Bio             *string   `db:"bio" json:"bio,omitempty"`
	ProfileImageURL *string   `db:"profile_image_url" json:"profile_image_url,omitempty"`
	IsPrivate       bool      `db:"is_private" json:"is_private"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
