package types

import "time"

type Role struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`

	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoleStore interface {
	GetRoles() ([]*Role, error)
	GetRoleByID(id string) (*Role, error)
	GetRoleByName(name string) (*Role, error)
	CreateRole(Role) error
	UpdateRole(id string, r Role) error
	DeleteRole(id string) error
}

type SetPayloadRole struct {
	Name string `form:"name" validate:"required,min=3"`
}

type SetPayloadUpdateRole struct {
	Name string `form:"name" validate:"omitempty,required,min=3"`
}

// relation many to many with users.
type Roles []Role
