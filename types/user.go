package types

import (
	"context"
	"time"
)

type User struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
	Roles     Roles     `json:"roles"`

	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Avatar   string `json:"avatar"`

	TokenVersion int `json:"token_version"`
}

type UserStore interface {
	GetUsersWithPagination(ctx context.Context, page int) ([]*User, int64, error)
	GetUsersForSearch(ctx context.Context) []*User

	GetUserWithRolesByID(ctx context.Context, id string) (*User, error)
	GetUserWithRolesByEmail(ctx context.Context, email string) (*User, error)

	CreateUser(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, id string, u *User) error
	DeleteUser(ctx context.Context, id string) error

	IncrementTokenVersion(ctx context.Context, id, token string) error
}

type SetPayloadLogin struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required"`
}

type SetPayloadUser struct {
	Name     string `form:"name" validate:"required,min=3"`
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=6"`
}

type SetPayloadUpdateUser struct {
	Name     string `form:"name" validate:"omitempty,required,min=3"`
	Email    string `form:"email" validate:"omitempty,required,email"`
	Password string `form:"password" validate:"omitempty,required,min=6"`
}
