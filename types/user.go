package types

import (
	"context"
	"time"
)

type User struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
	Roles     []Role    `json:"roles"`

	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"-"`
	Avatar      string `json:"avatar"`
	R2AvatarURL string `json:"r2_avatar_url"`

	TokenVersion int `json:"token_version"`
}

type UsersCachePage struct {
	Users []*User `json:"users"`

	LastPage int64 `json:"last_page"`
}

type UserStore interface {
	GetUsersWithPagination(ctx context.Context, page int) ([]*User, int64, error)
	GetUsersForSearch(ctx context.Context) []*User

	GetUserWithRolesByID(ctx context.Context, id string) (*User, error)
	GetUserWithRolesByIDWithoutCache(ctx context.Context, id string) (*User, error)
	GetUserWithRolesByEmail(ctx context.Context, email string) (*User, error)

	CreateUser(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, id string, u *User) error
	DeleteUser(ctx context.Context, id string) error

	IncrementTokenVersion(ctx context.Context, id, token string) error
}

type SetPayloadJSONLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SetPayloadJSONRegister struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,lowercase,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type SetPayloadUser struct {
	Name     string `form:"name" validate:"required,min=3"`
	Email    string `form:"email" validate:"required,lowercase,email"`
	Password string `form:"password" validate:"required,min=6"`
}

type SetPayloadUpdateUser struct {
	Name     string `form:"name" validate:"omitempty,required,min=3"`
	Email    string `form:"email" validate:"omitempty,required,lowercase,email"`
	Password string `form:"password" validate:"omitempty,required,min=6"`
}
