package types

import (
	"context"
	"time"
)

type Role struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`

	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoleStore interface {
	GetRoles(ctx context.Context) ([]*Role, error)

	GetRoleByID(ctx context.Context, id string) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)

	CreateRole(ctx context.Context, r Role) error
	UpdateRole(ctx context.Context, id string, r Role) error
	DeleteRole(ctx context.Context, id string) error
}

type SetPayloadRole struct {
	Name string `form:"name" validate:"required,min=3"`
}

type SetPayloadUpdateRole struct {
	Name string `form:"name" validate:"omitempty,required,min=3"`
}

// relation many to many with users.
type Roles []Role
