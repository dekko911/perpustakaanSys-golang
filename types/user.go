package types

import (
	"time"
)

type User struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
	Roles     Roles     `json:"roles"`

	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"-"`
	Avatar       string `json:"avatar"`
	TokenVersion int    `json:"token_version"`
}

type UserStore interface {
	GetUsers() ([]*User, error)
	GetUserWithRolesByID(id string) (*User, error)
	GetUserWithRolesByEmail(email string) (*User, error)
	CreateUser(*User) error
	UpdateUser(id string, u *User) error
	DeleteUser(id string) error
	IncrementTokenVersion(id string) error
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
